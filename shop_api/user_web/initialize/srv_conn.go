package initialize

import (
	"fmt"
	"shop/shop_api/user_web/global"
	"shop/shop_api/user_web/proto"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSrvConn() {

}

func InitSrvConn2() {
	// 从注册中心获取user_srv的信息
	// 设置consul信息获取client
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		zap.S().Fatalf("[InitSrvConn] 连接【consul】失败", "msg", err.Error())
		return
	}

	// 根据服务名查询服务ip+port
	var userSrvAddr string
	var userSrvPort int
	serviceName := global.ServerConfig.UserSrvCfg.Name
	zap.S().Infow("[InitSrvConn] 正在从Consul查询服务", "service_name", serviceName)

	data, err := client.Agent().ServicesWithFilter(fmt.Sprintf(`Service == "%s"`, serviceName))
	if err != nil {
		zap.S().Fatalf("[InitSrvConn] 获取【用户服务】地址失败", "msg", err.Error())
		return
	}

	// 从 Consul 获取服务地址
	if len(data) == 0 {
		zap.S().Fatalf("[InitSrvConn] 从Consul未找到服务", "service", serviceName)
		return
	}

	// 记录所有找到的服务实例
	zap.S().Infow("[InitSrvConn] 从Consul找到服务实例", "count", len(data))
	for id, v := range data {
		zap.S().Infow("[InitSrvConn] 服务实例详情", "service_id", id, "address", v.Address, "port", v.Port, "tags", v.Tags)
	}

	// 获取第一个匹配的服务实例
	found := false
	for _, v := range data {
		userSrvAddr = v.Address
		userSrvPort = v.Port
		zap.S().Infow("[InitSrvConn] 选择服务实例", "address", userSrvAddr, "port", userSrvPort, "service_id", v.ID)
		found = true
		break
	}

	if !found || userSrvAddr == "" {
		zap.S().Fatalf("[InitSrvConn] Consul中的服务地址无效", "address", userSrvAddr, "port", userSrvPort)
		return
	}

	// 与服务建立连接
	userSrvEndpoint := fmt.Sprintf("%s:%d", userSrvAddr, userSrvPort)
	zap.S().Infow("[InitSrvConn] 正在连接用户服务", "endpoint", userSrvEndpoint)
	conn, err := grpc.NewClient(userSrvEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Fatalf("[InitSrvConn] 连接【用户服务】失败", "msg", err.Error())
		return
	}

	// 创建rpc用户服务客户端
	global.UserSrvClient = proto.NewUserClient(conn)
}
