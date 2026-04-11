package memory

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/rahulbalajee/Movie/pkg/discovery"
)

const (
	defaultCutoff = 5 * time.Second
)

type Registry struct {
	mu           sync.RWMutex
	serviceAddrs map[string]map[string]*serviceInstance
	cutoff       time.Duration
}

type Options func(*Registry)

func WithTTL(ttl time.Duration) Options {
	return func(r *Registry) {
		r.cutoff = ttl
	}
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

func NewRegistry(opts ...Options) *Registry {
	r := &Registry{
		serviceAddrs: map[string]map[string]*serviceInstance{},
		cutoff:       defaultCutoff,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Registry) Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.serviceAddrs[serviceName]; !ok {
		r.serviceAddrs[serviceName] = map[string]*serviceInstance{}
	}

	r.serviceAddrs[serviceName][instanceId] = &serviceInstance{
		hostPort:   hostPort,
		lastActive: time.Now(),
	}
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instanceId string, serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return nil
	}

	delete(r.serviceAddrs[serviceName], instanceId)
	return nil
}

func (r *Registry) ReportHealthState(ctx context.Context, instanceId string, serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return errors.New("service is not registered yet")
	}

	if _, ok := r.serviceAddrs[serviceName][instanceId]; !ok {
		return errors.New("instance " + instanceId + " of service " + serviceName + " is not registered yet")
	}

	r.serviceAddrs[serviceName][instanceId].lastActive = time.Now()
	return nil
}

func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.serviceAddrs[serviceName]
	if !ok {
		return nil, errors.New("service is not registered yet")
	}
	if len(instances) == 0 {
		return nil, discovery.ErrNotFound
	}

	var res []string

	cutoff := time.Now().Add(-r.cutoff)

	for i, instance := range instances {
		if instance.lastActive.Before(cutoff) {
			log.Println("instance " + i + " of service " + serviceName + " is no longer active, skipping")
			continue
		}

		res = append(res, instance.hostPort)
	}

	if len(res) == 0 {
		return nil, discovery.ErrNotFound
	}

	return res, nil
}
