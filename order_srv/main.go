package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"shop/order_srv/global"
	"shop/order_srv/handler"
	"shop/order_srv/initialize"
	"shop/order_srv/proto"
	"shop/order_srv/utils/register/consul"
	"syscall"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()

	// 打印配置信息以便调试
	zap.S().Infow("服务配置", "name", global.ServerConfig.Name, "host", global.ServerConfig.Host, "port", global.ServerConfig.Port)
	initialize.InitDB()
	// 初始化分布式锁
	initialize.InitRedisSync()

	server := grpc.NewServer()
	// 注册用户服务
	proto.RegisterOrderServer(server, &handler.OrderServer{})
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", global.ServerConfig.Host, global.ServerConfig.Port))
	zap.S().Infof("server run at port %s:%d", global.ServerConfig.Host, global.ServerConfig.Port)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 向consul注册服务
	cfg := api.DefaultConfig()
	// consul server的地址
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// 向consul注册服务
	serviceId := uuid.NewString()
	c := consul.NewRegistry(global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	c.Register(global.ServerConfig.Host, global.ServerConfig.Port, global.ServerConfig.Name, []string{"order-srv"}, serviceId)

	// 启动grpc Server
	go func() {
		if err := server.Serve(listener); err != nil {
			zap.S().Fatal("grpc serve error: ", err)
		}
	}()
	zap.S().Info("gRPC server started")

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 停止grpc
	server.GracefulStop()
	// 注销旧的服务注册
	if err := client.Agent().ServiceDeregister(serviceId); err != nil {
		zap.S().Warnw("注销旧服务注册失败（可能服务不存在）", "service_id", serviceId, "error", err.Error())
	}
	zap.S().Infow("成功注销旧服务注册", "service_id", serviceId)
}
