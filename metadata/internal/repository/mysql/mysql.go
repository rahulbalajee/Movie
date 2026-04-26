package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rahulbalajee/Movie/metadata/internal/repository"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	var title string
	var desc, director sql.NullString
	row := r.db.QueryRowContext(
		ctx,
		`SELECT title, description, director FROM movies WHERE id = ?`,
		id,
	)

	if err := row.Scan(&title, &desc, &director); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &model.Metadata{
		ID:          id,
		Title:       title,
		Description: desc.String,
		Director:    director.String,
	}, nil
}

func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO movies (id, title, description, director) VALUES (?, ?, ?, ?)
         ON DUPLICATE KEY UPDATE title = VALUES(title), description = VALUES(description), director = VALUES(director)`,
		id, metadata.Title, metadata.Description, metadata.Director,
	)

	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
