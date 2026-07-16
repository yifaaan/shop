package registry

import (
	"fmt"
	"shop/services/user_web/config"

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

// Resolve returns the "host:port" of one healthy instance of the named
// service from Consul. Use it to discover downstream services instead of
// hardcoding their addresses. Only passing (healthy) instances are
// considered; the first one is returned.
func (r *Registrar) Resolve(name string) (string, error) {
	entries, _, err := r.client.Health().Service(name, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("query consul for %q: %w", name, err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instance of %q in consul", name)
	}
	entry := entries[0]
	addr := entry.Service.Address
	if addr == "" {
		addr = entry.Node.Address
	}
	return fmt.Sprintf("%s:%d", addr, entry.Service.Port), nil
}
