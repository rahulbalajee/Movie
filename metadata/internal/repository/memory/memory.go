package memory

import (
	"context"
	"sync"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

// Repository defines an in-memory movie metadata repository
type Repository struct {
	mu   sync.RWMutex
	data map[string]*model.Metadata
}

// Factory to create the repository
func NewRepo() *Repository {
	return &Repository{
		data: make(map[string]*model.Metadata),
	}
}

// Get retrives the movie metadata by movie id
func (r *Repository) Get(_ context.Context, id string) (*model.Metadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Put adds movie metadata for a given movie id
func (r *Repository) Put(_ context.Context, id string, metadata *model.Metadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[id] = metadata
	return nil
}
