package grpc

import (
	"context"
	"errors"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedMetadataServiceServer
	ctrl *metadata.Controller
}

func NewHandler(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetMetadata(ctx context.Context, req *gen.GetMetadataRequest) (*gen.GetMetadataResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}

	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil {
		if errors.Is(err, metadata.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &gen.GetMetadataResponse{Metadata: model.MetadataToProto(m)}, nil
}

func (h *Handler) PutMetadata(ctx context.Context, req *gen.PutMetadataRequest) (*gen.PutMetadataResponse, error) {
	if req == nil || req.Metadata == nil || req.Metadata.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty metadata id")
	}

	if err := h.ctrl.Put(ctx, req.Metadata.Id, model.MetadataFromProto(req.Metadata)); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &gen.PutMetadataResponse{}, nil
}
