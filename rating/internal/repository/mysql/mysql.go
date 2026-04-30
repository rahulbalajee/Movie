package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rahulbalajee/Movie/rating/internal/repository"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
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

func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?`,
		recordID,
		recordType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int

		if err := rows.Scan(&userID, &value); err != nil {
			return nil, err
		}

		res = append(res, model.Rating{
			UserID: model.UserID(userID),
			Value:  model.RatingValue(value),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, repository.ErrNotFound
	}

	return res, nil
}

func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	if rating == nil {
		return errors.New("rating is nil")
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE value = VALUES(value)`,
		recordID, recordType, rating.UserID, rating.Value,
	)

	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
