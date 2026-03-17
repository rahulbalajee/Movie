package metadata

import (
	"context"
	"errors"
	"fmt"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("not found")

// metadataRepository implements the database repository on the reciever side
type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
}

// Controller defines a metadata service controller
type Controller struct {
	repo metadataRepository
}

// Factory function to create a new controller
func NewController(repo metadataRepository) *Controller {
	return &Controller{repo: repo}
}

// Get returns movie metadata by id
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error getting movie record from repo %s: %w", id, err)
	}
	return res, nil
}
