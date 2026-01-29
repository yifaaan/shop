package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"shop/good_srv/global"
	"shop/good_srv/handler"
	"shop/good_srv/initialize"
	"shop/good_srv/proto"
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
	// 使用系统分配的端口

	// 打印配置信息以便调试
	zap.S().Infow("服务配置", "name", global.ServerConfig.Name, "host", global.ServerConfig.Host, "port", global.ServerConfig.Port)
	initialize.InitDB()

	server := grpc.NewServer()
	// 注册用户服务
	proto.RegisterGoodServer(server, &handler.GoodServer{})
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

	// 服务注册地址：供其他服务访问使用
	serviceAddress := global.ServerConfig.Host
	// 健康检查地址：如果 Consul 在 Docker 容器中，需要使用 host.docker.internal 来访问主机服务
	healthCheckAddr := global.ServerConfig.Host
	// 如果服务地址是 127.0.0.1，且 Consul 可能在容器中，使用 host.docker.internal 进行健康检查
	if serviceAddress == "127.0.0.1" || serviceAddress == "localhost" {
		healthCheckAddr = "host.docker.internal"
		zap.S().Infow("健康检查使用 host.docker.internal", "reason", "Consul可能在Docker容器中")
	}

	serviceId := uuid.New().String()
	reg := api.AgentServiceRegistration{
		ID:      serviceId,
		Name:    global.ServerConfig.Name,
		Tags:    []string{"good-srv"},
		Port:    global.ServerConfig.Port,
		Address: serviceAddress, // 服务地址：供其他服务访问使用
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", healthCheckAddr, global.ServerConfig.Port),
			Timeout:                        "5s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "10s",
		},
	}

	zap.S().Infow("服务注册配置",
		"service_address", serviceAddress,
		"health_check_address", healthCheckAddr,
		"port", global.ServerConfig.Port)

	err = client.Agent().ServiceRegister(&reg)
	if err != nil {
		panic(fmt.Sprintf("注册服务到Consul失败: %v", err))
	}
	zap.S().Infow("服务已注册到Consul", "service_id", serviceId, "address", global.ServerConfig.Host, "port", global.ServerConfig.Port)

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
