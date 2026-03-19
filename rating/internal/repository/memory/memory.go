package memory

import (
	"context"

	"github.com/rahulbalajee/Movie/rating/internal/repository"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

// Repository defines a rating repository
type Repository struct {
	data map[model.RecordType]map[model.RecordID][]model.Rating
}

// Factory to create the repository
func NewRepo() *Repository {
	return &Repository{
		data: make(map[model.RecordType]map[model.RecordID][]model.Rating),
	}
}

// Get retrieves all ratings for a given record
func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	if _, ok := r.data[recordType]; !ok {
		return nil, repository.ErrNotFound
	}

	ratings, ok := r.data[recordType][recordID]
	if !ok || len(ratings) == 0 {
		return nil, repository.ErrNotFound
	}

	return ratings, nil
}

// Put adds rating for a given record
func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	if _, ok := r.data[recordType]; !ok {
		r.data[recordType] = map[model.RecordID][]model.Rating{}
	}

	r.data[recordType][recordID] = append(r.data[recordType][recordID], *rating)
	return nil
}
