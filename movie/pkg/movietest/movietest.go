// Package movietest exposes constructors that wire up the movie service's
// dependency graph (gateways + controller + gRPC handler) in a single call.
// Intended for integration tests where the movie service must reach a
// real (in-process) metadata and rating service via the supplied registry.
package movietest

import (
	metadatagateway "github.com/rahulbalajee/Movie/movie/internal/gateway/metadata/grpc"
	ratinggateway "github.com/rahulbalajee/Movie/movie/internal/gateway/rating/grpc"
	"github.com/rahulbalajee/Movie/movie/internal/controller/movie"
	grpchandler "github.com/rahulbalajee/Movie/movie/internal/handler/grpc"
	"github.com/rahulbalajee/Movie/pkg/discovery"
)

// NewTestMovieGRPCServer returns a fully wired movie gRPC handler. Unlike
// the metadata and rating test servers, this one needs a discovery.Registry
// because the movie service has no local store — it fans out to metadata
// and rating via gRPC, and the gateways resolve those addresses through
// the registry.
func NewTestMovieGRPCServer(registry discovery.Registry) *grpchandler.Handler {
	metadataGw := metadatagateway.NewGateway(registry)
	ratingGw := ratinggateway.NewGateway(registry)
	ctrl := movie.NewController(ratingGw, metadataGw)
	return grpchandler.NewHandler(ctrl)
}
