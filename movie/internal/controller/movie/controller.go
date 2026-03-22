package movie

import (
	"context"
	"errors"

	metadatamodel "github.com/rahulbalajee/Movie/metadata/pkg/model"
	"github.com/rahulbalajee/Movie/movie/internal/gateway"
	"github.com/rahulbalajee/Movie/movie/pkg/model"
	ratingmodel "github.com/rahulbalajee/Movie/rating/pkg/model"
)

// ErrNotFound is returned when movie metadata is not found
var ErrNotFound = errors.New("movie metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordId ratingmodel.RecordID, recordType ratingmodel.RecordType) (float64, error)
	PutRating(ctx context.Context, recordId ratingmodel.RecordID, recordType ratingmodel.RecordType, rating *ratingmodel.Rating) error
}

type metadataGateway interface {
	Get(ctx context.Context, id string) (*metadatamodel.Metadata, error)
}

// Controller defines a movie service controller
type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

// New creates a new movie service controller
func NewController(ratingGateway ratingGateway, metadataGateway metadataGateway) *Controller {
	return &Controller{
		ratingGateway:   ratingGateway,
		metadataGateway: metadataGateway,
	}
}

// Get returns the movie details including the aggregated ratings and movie metadata
func (c *Controller) Get(ctx context.Context, id string) (*model.MovieDetails, error) {
	metadata, err := c.metadataGateway.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gateway.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	details := &model.MovieDetails{
		Metadata: *metadata,
	}

	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordID(id), ratingmodel.RecordTypeMovie)
	if err != nil {
		if errors.Is(err, gateway.ErrNotFound) {
			// Just proceed in this case, it's ok not to have ratings yet
		}
		return nil, err
	}
	details.Rating = &rating

	return details, nil
}
