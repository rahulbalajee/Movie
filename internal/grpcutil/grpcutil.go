package grpcutil

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/rahulbalajee/Movie/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error getting service addresses: %w", err)
	}

	return grpc.NewClient(addrs[rand.IntN(len(addrs))], grpc.WithTransportCredentials(insecure.NewCredentials()))
}
