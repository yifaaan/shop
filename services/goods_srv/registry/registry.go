package registry

import (
	"fmt"

	"shop/services/goods_srv/config"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

// Registrar wraps a Consul agent client for one service registration.
type Registrar struct {
	client    *api.Client
	serviceID string
}

// New registers the gRPC service described by cfg with the local Consul agent.
// The check uses Consul's native gRPC health check (grpc.health.v1).
func New(cfg *config.Config) (*Registrar, error) {
	client, err := api.NewClient(&api.Config{
		Address: fmt.Sprintf("%s:%d", cfg.Consul.Host, cfg.Consul.Port),
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	// 服务名相同供发现；ID 用 UUID 唯一，多实例才不会在 Consul 互相覆盖。
	// 进程崩溃后健康检查会失败，DeregisterCriticalServiceAfter 让 Consul 自动清掉残留实例。
	serviceID := uuid.NewString()
	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Name,
		Address: cfg.Consul.Address,
		Port:    cfg.Port,
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d/%s", cfg.Consul.Address, cfg.Port, cfg.Name),
			GRPCUseTLS:                     false,
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("register service %q: %w", cfg.Name, err)
	}
	return &Registrar{client: client, serviceID: serviceID}, nil
}

// Deregister removes the service from the local Consul agent.
func (r *Registrar) Deregister() error {
	return r.client.Agent().ServiceDeregister(r.serviceID)
}