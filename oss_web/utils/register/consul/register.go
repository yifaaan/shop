package consul

import (
	"fmt"
	"shop/oss_web/global"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type RegistryClient interface {
	Register(address string, port int, name string, tags []string, id string)
	DeRegister(id string) error
}

var _ RegistryClient = (*Registry)(nil)

type Registry struct {
	Host string
	Port int
}

func NewRegistry(host string, port int) RegistryClient {
	return &Registry{
		Host: host,
		Port: port,
	}
}

func (r *Registry) Register(address string, port int, name string, tags []string, id string) {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// 服务注册地址：供其他服务访问使用
	serviceAddress := address
	// 健康检查地址：如果 Consul 在 Docker 容器中，需要使用 host.docker.internal 来访问主机服务
	healthCheckAddr := address
	// 如果服务地址是 127.0.0.1，且 Consul 可能在容器中，使用 host.docker.internal 进行健康检查
	if serviceAddress == "127.0.0.1" || serviceAddress == "localhost" {
		healthCheckAddr = "host.docker.internal"
		zap.S().Infow("健康检查使用 host.docker.internal", "reason", "Consul可能在Docker容器中")
	}

	reg := api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Tags:    tags,
		Port:    port,
		Address: serviceAddress, // 服务地址：供其他服务访问使用
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/g/v1/health", healthCheckAddr, port),
			Timeout:                        "5s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "10s",
		},
	}

	zap.S().Infow("服务注册配置",
		"service_address", serviceAddress,
		"health_check_address", healthCheckAddr,
		"port", port)

	err = client.Agent().ServiceRegister(&reg)
	if err != nil {
		panic(fmt.Sprintf("注册服务到Consul失败: %v", err))
	}
	zap.S().Infow("服务已注册到Consul", "service_id", id, "address", global.ServerConfig.IP, "port", global.ServerConfig.Port)
}

func (r *Registry) DeRegister(id string) error {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}
	err = client.Agent().ServiceDeregister(id)
	return err
}
