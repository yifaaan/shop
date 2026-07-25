package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop/pkg/port"
	"shop/pkg/proto"
	"shop/services/order_web/config"
	"shop/services/order_web/registry"
	"shop/services/order_web/web"

	_ "github.com/mbobakov/grpc-consul-resolver" // 注册 "consul" gRPC resolver scheme
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	if err := web.ConfigureTranslator("zh"); err != nil {
		log.Panic("failed to init translator: ", err)
	}

	// 注册到 Consul
	reg, err := registry.New(cfg)
	if err != nil {
		log.Panic("failed to register with consul: ", err)
	}

	// 通过 Consul resolver 发现 order_srv 并按 round_robin 负载均衡。
	orderConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true", cfg.Consul.Host, cfg.Consul.Port, cfg.OrderSrv.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Panic("failed to connect to order service: ", err)
	}
	defer orderConn.Close()

	// 通过 Consul resolver 发现 goods_srv（购物车列表需补全商品名/图/价）。
	goodsConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true", cfg.Consul.Host, cfg.Consul.Port, cfg.GoodsSrv.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Panic("failed to connect to goods service: ", err)
	}
	defer goodsConn.Close()

	// 通过 Consul resolver 发现 inventory_srv（加购/下单需校验库存）。
	inventoryConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true", cfg.Consul.Host, cfg.Consul.Port, cfg.InventorySrv.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Panic("failed to connect to inventory service: ", err)
	}
	defer inventoryConn.Close()

	srv := web.New(cfg, log, web.NewJWT(cfg), proto.NewOrderClient(orderConn), proto.NewGoodsClient(goodsConn), proto.NewInventoryClient(inventoryConn))
	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           srv.Routers(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Infof("starting server, port: %d", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panic("failed to start server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error: ", err)
	}

	if err := reg.Deregister(); err != nil {
		log.Error("deregister from consul error: ", err)
	}
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
