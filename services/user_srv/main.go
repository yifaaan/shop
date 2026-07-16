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

	"shop/pkg/proto"
	"shop/services/user_srv/config"
	"shop/services/user_srv/registry"

	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	db, err := openDB(cfg.MySQL, log)
	if err != nil {
		log.Panic("failed to open db: ", err)
	}
	log.Info("DB init done")

	server := grpc.NewServer()
	proto.RegisterUserServer(server, NewUserServer(db))

	// gRPC 标准健康检查服务（grpc.health.v1）
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
		// 先标 NOT_SERVING，让 Consul 健康检查立刻收到失败，再优雅停机
		healthSrv.SetServingStatus(cfg.Name, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		server.GracefulStop()
		if err := reg.Deregister(); err != nil {
			log.Error("deregister from consul error: ", err)
		}
	}
}

// openDB opens the MySQL connection and auto-migrates the User table.
// gorm log output is routed through the injected zap logger via zapWriter.
func openDB(cfg config.MySQLConfig, log *zap.SugaredLogger) (*gorm.DB, error) {
	gormLogger := logger.New(
		zapWriter{log: log},
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn, // 只记慢查询(>SlowThreshold)和错误，不打印每条 SQL
			Colorful:      false,
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
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

func (w zapWriter) Printf(format string, args ...interface{}) {
	w.log.Info(strings.TrimRight(fmt.Sprintf(format, args...), "\n"))
}
