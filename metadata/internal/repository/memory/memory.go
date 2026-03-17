package memory

import (
	"context"
	"sync"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg"
)

// Repository defines an in-memory movie metadata repository
type Repository struct {
	sync.RWMutex
	data map[string]*pkg.Metadata
}

// Factory to create the repository
func NewRepo() *Repository {
	return &Repository{
		data: make(map[string]*pkg.Metadata),
	}
}

// Get retrives the movie metadata by movie id
func (r *Repository) Get(_ context.Context, id string) (*pkg.Metadata, error) {
	r.RLock()
	defer r.RUnlock()

	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

func (r *Repository) Put(_ context.Context, id string, metadata *pkg.Metadata) error {
	r.Lock()
	defer r.Unlock()

	r.data[id] = metadata
	return nil
}
