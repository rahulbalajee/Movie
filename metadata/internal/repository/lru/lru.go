// Package lru provides a bounded, TTL-aware in-memory cache for movie
// metadata. It satisfies the same Get/Put contract as the primary
// repository so it can be used wherever the controller expects a
// metadataRepository.
package lru

import (
	"context"
	"time"

	expirable "github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

// Cache is an LRU cache with per-entry TTL. It is safe for concurrent use.
type Cache struct {
	c *expirable.LRU[string, *model.Metadata]
}

// New returns a Cache that holds at most `size` entries and expires each
// entry after `ttl`. A size of 0 disables the size bound; a ttl of 0
// disables expiry.
func New(size int, ttl time.Duration) *Cache {
	return &Cache{
		c: expirable.NewLRU[string, *model.Metadata](size, nil, ttl),
	}
}

// Get returns the cached metadata for id, or repository.ErrNotFound on a
// miss (including expired entries).
func (c *Cache) Get(_ context.Context, id string) (*model.Metadata, error) {
	m, ok := c.c.Get(id)
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Put inserts metadata for id, evicting the least-recently-used entry if
// the cache is at capacity.
func (c *Cache) Put(_ context.Context, id string, metadata *model.Metadata) error {
	c.c.Add(id, metadata)
	return nil
}

// Delete removes the entry for id, if present. Used by the controller to
// invalidate the cache after a write so the next reader sees fresh data
// instead of waiting for the TTL.
func (c *Cache) Delete(_ context.Context, id string) error {
	c.c.Remove(id)
	return nil
}
