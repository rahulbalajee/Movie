package rating

import (
	"context"
	"errors"
	"fmt"

	"github.com/rahulbalajee/Movie/rating/internal/repository"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

// ErrNotFound is returned when ratings are not found for a record
var ErrNotFound = errors.New("ratings not found for a record")

// ratingRepository implements the rating repository on the reciever side
type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

type ratingIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

// Controller defines the rating service controller
type Controller struct {
	repo     ratingRepository
	ingester ratingIngester
}

// Factory function to create a new controller
func NewController(repo ratingRepository, ingester ratingIngester) *Controller {
	return &Controller{
		repo:     repo,
		ingester: ingester,
	}
}

func (c *Controller) StartIngestion(ctx context.Context) error {
	ch, err := c.ingester.Ingest(ctx)
	if err != nil {
		return err
	}

	for e := range ch {
		fmt.Printf("consumed a message: %v\n", e)
		if err := c.PutRating(
			ctx,
			e.RecordID,
			e.RecordType,
			&model.Rating{
				UserID: e.UserID,
				Value:  e.Value,
			},
		); err != nil {
			return err
		}
	}

	return nil
}

// GetAggregatedRating returns aggregated rating for a record or ErrNotFound in case no ratings exists for that record
func (c *Controller) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordID, recordType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

// PutRating writes a rating for a given record
func (c *Controller) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordID, recordType, rating)
}
