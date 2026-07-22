package registry

import (
	"fmt"

	"shop/services/goods_web/config"

	"github.com/hashicorp/consul/api"
)

// Registrar wraps a Consul agent client for one service registration.
type Registrar struct {
	client    *api.Client
	serviceID string
}

func New(cfg *config.Config) (*Registrar, error) {
	client, err := api.NewClient(&api.Config{
		Address: fmt.Sprintf("%s:%d", cfg.Consul.Host, cfg.Consul.Port),
	})
	if err != nil {
		return nil, fmt.Errorf("create Consul client: %w", err)
	}
	// 单实例用服务名作 ID；多实例时改 name + "-" + hostname 唯一化。
	serviceID := cfg.Name
	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Name,
		Address: cfg.Consul.Address,
		Port:    cfg.Port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", cfg.Consul.Address, cfg.Port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}
	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("register service %q: %w", cfg.Name, err)
	}
	return &Registrar{
		client:    client,
		serviceID: serviceID,
	}, nil
}

// Deregister removes the service from the local Consul agent.
func (r *Registrar) Deregister() error {
	return r.client.Agent().ServiceDeregister(r.serviceID)
}