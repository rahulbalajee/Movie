package grpc

import (
	"context"
	"errors"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
	"github.com/rahulbalajee/Movie/movie/internal/controller/movie"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMovieServiceServer
	ctrl *movie.Controller
}

func NewHandler(ctrl *movie.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}

	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil {
		if errors.Is(err, movie.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	resp := &gen.GetMovieDetailsResponse{
		MovieDetails: &gen.Movie{
			Metadata: model.MetadataToProto(&m.Metadata),
		},
	}
	// Rating is optional — a movie with no ratings yet has m.Rating == nil.
	if m.Rating != nil {
		resp.MovieDetails.Rating = *m.Rating
	}
	return resp, nil
}
