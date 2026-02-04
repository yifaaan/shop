package initialize

import (
	"fmt"
	"shop/order_srv/global"
	"shop/order_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitGoodSrvConn() {

	// grpc-consul-resolver进程内负载均衡
	consulAddr := fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port, global.ServerConfig.GoodSrvCfg.Name)
	conn, err := grpc.NewClient(
		consulAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("new grpc client failed: %v", err)
	}
	global.GoodSrvClient = proto.NewGoodClient(conn)
}

func InitInventorySrvConn() {

	// grpc-consul-resolver进程内负载均衡
	consulAddr := fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port, global.ServerConfig.InventorySrvCfg.Name)
	conn, err := grpc.NewClient(
		consulAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("new grpc client failed: %v", err)
	}
	global.InventorySrvClient = proto.NewInventoryClient(conn)
}
