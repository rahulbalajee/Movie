package metadata

//go:generate mockgen -source=controller.go -destination=mocks_test.go -package=metadata

import (
	"context"
	"errors"
	"fmt"
	"log"

	"golang.org/x/sync/singleflight"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("not found")

// metadataRepository implements the database repository on the reciever side
type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
	Delete(ctx context.Context, id string) error
}

// Controller defines a metadata service controller
type Controller struct {
	repo  metadataRepository
	cache metadataRepository
	// sf collapses concurrent cache misses for the same id into a single
	// downstream Get against repo, so a thundering herd of N callers only
	// produces 1 backend request.
	sf singleflight.Group
}

// Factory function to create a new controller
func NewController(repo metadataRepository, cache metadataRepository) *Controller {
	return &Controller{
		repo:  repo,
		cache: cache,
	}
}

// Get returns movie metadata by id
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	if cacheRes, err := c.cache.Get(ctx, id); err == nil {
		return cacheRes, nil
	}

	v, err, _ := c.sf.Do(id, func() (any, error) {
		// Re-check the cache inside the singleflight to catch the case
		// where another goroutine populated it while we were queued.
		if cacheRes, err := c.cache.Get(ctx, id); err == nil {
			return cacheRes, nil
		}

		res, err := c.repo.Get(ctx, id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("error getting movie record from repo %s: %w", id, err)
		}

		if err := c.cache.Put(ctx, id, res); err != nil {
			log.Printf("failed to update the cache for %s: %v", id, err)
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*model.Metadata), nil
}

// Put writes metadata to the primary repository and invalidates the cache
// entry so the next Get observes the new value immediately rather than
// waiting up to the cache TTL.
//
// We invalidate (delete) instead of write-through to keep the cache as
// the single source of truth for "fresh-or-absent" — the next reader
// pays one miss and repopulates with whatever the repo actually has,
// which is robust against partial-write races.
func (c *Controller) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	if err := c.repo.Put(ctx, id, metadata); err != nil {
		return fmt.Errorf("error putting movie record %s: %w", id, err)
	}

	// Best-effort invalidation: a failure here only widens the staleness
	// window to the cache TTL, it does not corrupt data.
	if err := c.cache.Delete(ctx, id); err != nil {
		log.Printf("failed to invalidate cache for %s: %v", id, err)
	}

	// Forget any in-flight singleflight result for this key. Without this,
	// a Get that started before our Put could finish *after* the Delete
	// above and repopulate the cache with stale data fetched pre-write.
	c.sf.Forget(id)

	return nil
}
