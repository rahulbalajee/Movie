// Package ratingtest exposes constructors that wire up the rating
// service's dependency graph (memory repo + controller + gRPC handler) in
// a single call. Intended for integration tests that want a real in-process
// server without depending on MySQL or Kafka.
package ratingtest

import (
	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	grpchandler "github.com/rahulbalajee/Movie/rating/internal/handler/grpc"
	"github.com/rahulbalajee/Movie/rating/internal/repository/memory"
)

// NewTestRatingGRPCServer returns a fully wired rating gRPC handler backed
// by the in-memory repository. The Kafka ingester is intentionally nil —
// integration tests exercise only the synchronous gRPC paths
// (GetAggregatedRating / PutRating), neither of which touches the ingester.
// Calling StartIngestion on the returned server will panic.
func NewTestRatingGRPCServer() *grpchandler.Handler {
	repo := memory.NewRepo()
	ctrl := rating.NewController(repo, nil)
	return grpchandler.NewHandler(ctrl)
}
