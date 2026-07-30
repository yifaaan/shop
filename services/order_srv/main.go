package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"shop/pkg/port"
	"shop/pkg/proto"
	"shop/services/order_srv/config"
	"shop/services/order_srv/registry"

	_ "github.com/mbobakov/grpc-consul-resolver" // 注册 "consul" gRPC resolver scheme
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	log := newLogger(false)
	defer func() { _ = log.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		log.Panic("failed to load config: ", err)
	}
	log = newLogger(cfg.Debug)

	// DEBUG 用配置里的固定端口，否则用 OS 分配的动态端口
	cfg.Port, err = port.Get(cfg.Debug, cfg.Port)
	if err != nil {
		log.Panic("get listen port: ", err)
	}

	db, err := openDB(cfg.MySQL, log)
	if err != nil {
		log.Panic("failed to open db: ", err)
	}
	log.Info("DB init done")

	// 通过 Consul resolver 发现下游服务，round_robin 负载均衡。
	goodsConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true", cfg.Consul.Host, cfg.Consul.Port, cfg.GoodsSrv.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Panic("failed to connect to goods service: ", err)
	}
	defer goodsConn.Close()

	inventoryConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true", cfg.Consul.Host, cfg.Consul.Port, cfg.InventorySrv.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Panic("failed to connect to inventory service: ", err)
	}
	defer inventoryConn.Close()

	// 初始化事务消息生产者（订单超时归还库存）
	txProducer, err := newRocketMQProducer(cfg.RocketMQ, db, log)
	if err != nil {
		log.Panic("failed to init rocketmq producer: ", err)
	}

	server := grpc.NewServer()
	proto.RegisterOrderServer(server, NewOrderServer(db, proto.NewGoodsClient(goodsConn), proto.NewInventoryClient(inventoryConn), log, txProducer))

	// gRPC 标准健康检查服务
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthSrv)
	healthSrv.SetServingStatus(cfg.Name, grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}

	// 注册到 Consul gRPC check
	reg, err := registry.New(cfg)
	if err != nil {
		log.Panic("failed to register with consul: ", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Infof("starting grpc server on %s:%d", cfg.Host, cfg.Port)
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Panic("failed to start grpc: ", err)
	case <-ctx.Done():
		log.Info("shutting down grpc server")
		healthSrv.SetServingStatus(cfg.Name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		server.GracefulStop()
		if err := txProducer.Close(); err != nil {
			log.Error("close rocketmq producer error: ", err)
		}
		if err := reg.Deregister(); err != nil {
			log.Error("deregister from consul error: ", err)
		}
	}
}

var migrateModels = []any{
	&ShoppingCart{}, &OrderInfo{}, &OrderGoods{},
}

// gormConfig 是建库连接与业务库连接共用的 GORM 配置。
func gormConfig(log *zap.SugaredLogger) *gorm.Config {
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(
			zapWriter{log: log},
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Warn, // 只记慢查询(>SlowThreshold)和错误，不打印每条 SQL
				Colorful:      false,
			},
		),
	}
}

func ensureDatabase(cfg config.MySQLConfig, log *zap.SugaredLogger) error {
	db, err := gorm.Open(mysql.Open(cfg.DSNWithoutDB()), gormConfig(log))
	if err != nil {
		return fmt.Errorf("connect mysql server: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	stmt := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	)
	if err := db.Exec(stmt).Error; err != nil {
		return fmt.Errorf("create database %q: %w", cfg.DBName, err)
	}
	return nil
}

// openDB 连接到 MySQL，确保业务库存在后，自动迁移 order 服务所需的表。
func openDB(cfg config.MySQLConfig, log *zap.SugaredLogger) (*gorm.DB, error) {
	if err := ensureDatabase(cfg, log); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig(log))
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(migrateModels...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

// newLogger returns a dev (console, debug-level) or prod (json, info-level) sugared logger.
func newLogger(debug bool) *zap.SugaredLogger {
	var (
		l   *zap.Logger
		err error
	)
	if debug {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	return l.Sugar()
}

// zapWriter adapts a *zap.SugaredLogger to gorm's logger.Writer interface.
type zapWriter struct {
	log *zap.SugaredLogger
}

func (w zapWriter) Printf(format string, args ...any) {
	w.log.Info(strings.TrimRight(fmt.Sprintf(format, args...), "\n"))
}
