package initialize

import (
	"fmt"
	"shop/order_web/global"
	"shop/order_web/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSrvConn() {

	// grpc-consul-resolver进程内负载均衡
	consulAddr := fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port, global.ServerConfig.OrderSrvCfg.Name)
	conn, err := grpc.NewClient(
		consulAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("new grpc client failed: %v", err)
	}
	global.OrderSrvClient = proto.NewOrderClient(conn)

	consulAddr = fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port, global.ServerConfig.GoodSrvCfg.Name)
	conn, err = grpc.NewClient(
		consulAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("new grpc client failed: %v", err)
	}
	global.GoodSrvClient = proto.NewGoodClient(conn)
}
