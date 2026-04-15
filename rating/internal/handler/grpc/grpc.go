package grpc

import (
	"context"
	"errors"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedRatingServiceServer
	ctrl *rating.Controller
}

func NewHandler(ctrl *rating.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetAggregatedRatings(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id/type")
	}

	v, err := h.ctrl.GetAggregatedRatings(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType))
	if err != nil {
		if errors.Is(err, rating.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &gen.GetAggregatedRatingResponse{RatingValue: v}, nil
}

func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty userid/record id")
	}

	if err := h.ctrl.PutRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType), &model.Rating{UserID: model.UserID(req.UserId), Value: model.RatingValue(req.RatingValue)}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &gen.PutRatingResponse{}, nil
}
