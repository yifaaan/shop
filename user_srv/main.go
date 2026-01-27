package main

import (
	"fmt"
	"net"
	"shop/user_srv/global"
	"shop/user_srv/handler"
	"shop/user_srv/initialize"
	"shop/user_srv/proto"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	server := grpc.NewServer()
	// 注册用户服务
	proto.RegisterUserServer(server, &handler.UserServer{})
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

	reg := api.AgentServiceRegistration{
		ID:      global.ServerConfig.Name,
		Name:    global.ServerConfig.Name,
		Tags:    []string{"user-srv"},
		Port:    global.ServerConfig.Port,
		Address: "host.docker.internal", // consul对外公布的usr_srv服务的ip，供其他服务访问使用
		Check: &api.AgentServiceCheck{
			// GRPC:                           fmt.Sprintf("%s:%d", global.ServerConfig.Host, global.ServerConfig.Port),
			GRPC:                           "host.docker.internal:8080", // 容器内部会解析到主机IP，从而访问主机上运行的user_srv服务
			Timeout:                        "5s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "10s",
		},
	}

	err = client.Agent().ServiceRegister(&reg)
	if err != nil {
		panic(err)
	}

	err = server.Serve(listener)
	if err != nil {
		panic("failed to serve: " + err.Error())
	}
}
