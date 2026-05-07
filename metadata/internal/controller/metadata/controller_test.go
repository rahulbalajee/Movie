package metadata

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

func TestController_Get_CacheHit_SkipsRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	want := &model.Metadata{ID: "1", Title: "Inception"}
	cache.EXPECT().Get(gomock.Any(), "1").Return(want, nil)
	// repo.Get must NOT be called on a cache hit. gomock will fail the test
	// if any unexpected call is made, so the absence of an EXPECT() here
	// is itself the assertion.

	c := NewController(repo, cache)
	got, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestController_Get_CacheMiss_FetchesFromRepoAndPopulatesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	want := &model.Metadata{ID: "1", Title: "Inception"}

	gomock.InOrder(
		cache.EXPECT().Get(gomock.Any(), "1").Return(nil, errors.New("miss")),
		cache.EXPECT().Get(gomock.Any(), "1").Return(nil, errors.New("miss")),
		repo.EXPECT().Get(gomock.Any(), "1").Return(want, nil),
		cache.EXPECT().Put(gomock.Any(), "1", want).Return(nil),
	)

	c := NewController(repo, cache)
	got, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestController_Get_RepoNotFound_ReturnsControllerErrNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	cache.EXPECT().Get(gomock.Any(), "missing").Return(nil, errors.New("miss")).Times(2)
	repo.EXPECT().Get(gomock.Any(), "missing").Return(nil, repository.ErrNotFound)

	c := NewController(repo, cache)
	_, err := c.Get(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected controller.ErrNotFound, got %v", err)
	}
}

func TestController_Get_RepoError_WrapsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	repoErr := errors.New("db down")
	cache.EXPECT().Get(gomock.Any(), "1").Return(nil, errors.New("miss")).Times(2)
	repo.EXPECT().Get(gomock.Any(), "1").Return(nil, repoErr)

	c := NewController(repo, cache)
	_, err := c.Get(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestController_Get_CachePutFails_StillReturnsValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	want := &model.Metadata{ID: "1", Title: "Inception"}
	cache.EXPECT().Get(gomock.Any(), "1").Return(nil, errors.New("miss")).Times(2)
	repo.EXPECT().Get(gomock.Any(), "1").Return(want, nil)
	cache.EXPECT().Put(gomock.Any(), "1", want).Return(errors.New("cache offline"))

	c := NewController(repo, cache)
	got, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("cache write failure should not surface to caller, got: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestController_Put_WritesRepoAndInvalidatesCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	m := &model.Metadata{ID: "1", Title: "Inception"}
	gomock.InOrder(
		repo.EXPECT().Put(gomock.Any(), "1", m).Return(nil),
		cache.EXPECT().Delete(gomock.Any(), "1").Return(nil),
	)

	c := NewController(repo, cache)
	if err := c.Put(context.Background(), "1", m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestController_Put_RepoFails_DoesNotInvalidateCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	m := &model.Metadata{ID: "1", Title: "Inception"}
	repo.EXPECT().Put(gomock.Any(), "1", m).Return(errors.New("disk full"))
	// cache.Delete must NOT be called when the primary write fails — we
	// don't want to invalidate a cache entry that may still be the truth.

	c := NewController(repo, cache)
	if err := c.Put(context.Background(), "1", m); err == nil {
		t.Fatal("expected error from repo failure, got nil")
	}
}

func TestController_Put_CacheDeleteFails_StillReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockmetadataRepository(ctrl)
	cache := NewMockmetadataRepository(ctrl)

	m := &model.Metadata{ID: "1", Title: "Inception"}
	repo.EXPECT().Put(gomock.Any(), "1", m).Return(nil)
	cache.EXPECT().Delete(gomock.Any(), "1").Return(errors.New("cache offline"))

	c := NewController(repo, cache)
	if err := c.Put(context.Background(), "1", m); err != nil {
		t.Fatalf("cache invalidation failure should not surface to caller, got: %v", err)
	}
}
