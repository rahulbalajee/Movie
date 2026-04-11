package consul

import (
	"context"
	"fmt"
	"net"
	"strconv"

	capi "github.com/hashicorp/consul/api"
	"github.com/rahulbalajee/Movie/pkg/discovery"
)

const (
	defaultTTL = "5s"
)

type Registry struct {
	client *capi.Client
	ttl    string
}

type Option func(*Registry)

func WithTTL(ttl string) Option {
	return func(r *Registry) {
		r.ttl = ttl
	}
}

func NewRegistry(addr string, opts ...Option) (*Registry, error) {
	config := capi.DefaultConfig()
	config.Address = addr

	client, err := capi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating client from consul: %w", err)
	}

	r := &Registry{client: client, ttl: defaultTTL}
	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("error splitting hostport var: %w", err)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("error parsing port: %w", err)
	}

	return r.client.Agent().ServiceRegisterOpts(
		&capi.AgentServiceRegistration{
			Address: host,
			ID:      instanceId,
			Name:    serviceName,
			Port:    p,
			Check: &capi.AgentServiceCheck{
				CheckID: instanceId,
				TTL:     r.ttl,
			},
		},
		capi.ServiceRegisterOpts{}.WithContext(ctx),
	)
}

func (r *Registry) Deregister(ctx context.Context, instanceId string, _ string) error {
	return r.client.Agent().ServiceDeregisterOpts(
		instanceId,
		(&capi.QueryOptions{}).WithContext(ctx),
	)
}

func (r *Registry) ReportHealthyState(ctx context.Context, instanceId string, _ string) error {
	return r.client.Agent().UpdateTTLOpts(
		instanceId,
		"",
		capi.HealthPassing,
		(&capi.QueryOptions{}).WithContext(ctx),
	)
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(
		serviceName,
		"",
		true,
		(&capi.QueryOptions{}).WithContext(ctx),
	)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}

	res := []string{}
	for _, entry := range entries {
		res = append(res, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	return res, nil
}
