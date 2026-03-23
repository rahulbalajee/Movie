package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Registry defines a service registry
type Registry interface {
	// Register creates a service instance record in the registry
	Register(ctx context.Context, instanceId string, serviceName string, hostPort string) error
	// Deregister removes a service instance record from the registry
	Deregister(ctx context.Context, instanceId string, serviceName string) error
	// ServiceAddresses returns the list of addresses of active instances of a given service
	ServiceAddresses(ctx context.Context, serviceId string) ([]string, error)
	// ReportHealthState is a push mechanism for reporting healthy state to a registry
	ReportHealthyState(instanceId string, serviceName string) error
}

// ErrNotFound is returned when no services addresses are found
var ErrNotFound = errors.New("no service addresses found")

// GenerateInstanceId generates a pseudo random service instance identifier using the service name suffixed by dash and a random number
func GenerateInstanceId(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
