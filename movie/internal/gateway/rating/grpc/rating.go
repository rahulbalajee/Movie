package grpc

import (
	"context"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/internal/grpcutil"
	"github.com/rahulbalajee/Movie/movie/internal/gateway"
	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGateway(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (g *Gateway) GetAggregatedRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType) (float64, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)

	resp, err := client.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{RecordId: string(recordId), RecordType: string(recordType)})
	if err != nil {
		// Translate the gRPC status to our domain sentinel so callers can
		// use errors.Is(err, gateway.ErrNotFound) without depending on gRPC types.
		if status.Code(err) == codes.NotFound {
			return 0, gateway.ErrNotFound
		}
		return 0, err
	}

	return resp.RatingValue, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)

	_, err = client.PutRating(ctx, &gen.PutRatingRequest{UserId: string(rating.UserID), RecordId: string(recordId), RecordType: string(recordType), RatingValue: int32(rating.Value)})
	if err != nil {
		return err
	}

	return nil
}
