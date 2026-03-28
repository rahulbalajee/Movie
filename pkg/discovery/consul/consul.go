package consul

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	capi "github.com/hashicorp/consul/api"
	"github.com/rahulbalajee/Movie/pkg/discovery"
)

const defaultTTL = "5s"

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
		return nil, err
	}

	r := &Registry{client: client, ttl: defaultTTL}
	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("hostPort must be in a form of <host>:<port>, example: localhost:8081")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	return r.client.Agent().ServiceRegister(
		&capi.AgentServiceRegistration{
			Address: parts[0],
			ID:      instanceId,
			Name:    serviceName,
			Port:    port,
			Check: &capi.AgentServiceCheck{
				CheckID: instanceId,
				TTL:     r.ttl,
			},
		},
	)
}

func (r *Registry) Deregister(ctx context.Context, instanceId string, _ string) error {
	return r.client.Agent().ServiceDeregister(instanceId)
}

func (r *Registry) ReportHealthyState(instanceId string, _ string) error {
	return r.client.Agent().PassTTL(instanceId, "")
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, discovery.ErrNotFound
	}

	var res []string
	for _, entry := range entries {
		res = append(res, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	return res, nil
}
