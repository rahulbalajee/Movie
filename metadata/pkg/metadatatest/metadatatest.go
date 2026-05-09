// Package metadatatest exposes constructors that wire up the metadata
// service's dependency graph (memory repo + controller + gRPC handler) in
// a single call. Intended for integration tests that want a real in-process
// server without depending on MySQL or Consul.
package metadatatest

import (
	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	grpchandler "github.com/rahulbalajee/Movie/metadata/internal/handler/grpc"
	"github.com/rahulbalajee/Movie/metadata/internal/repository/memory"
)

// NewTestMetadataGRPCServer returns a fully wired metadata gRPC handler
// backed by the in-memory repository (used as both primary store and cache).
func NewTestMetadataGRPCServer() *grpchandler.Handler {
	repo := memory.NewRepo()
	cache := memory.NewRepo()
	ctrl := metadata.NewController(repo, cache)
	return grpchandler.NewHandler(ctrl)
}
