package internal

import (
	"context"
	"database/sql"
	"errors"
)

type PostgresURLRepository struct {
	db *sql.DB
}

type URLRepository interface {
	Save(ctx context.Context, original string) (int64, error)
	FindByID(ctx context.Context, id int64) (*URL, error)
	IncrementClicks(ctx context.Context, id int64) error
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func (r *PostgresURLRepository) Save(ctx context.Context, original string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, "INSERT INTO urls (original, clicks) VALUES ($1, 0) RETURNING id", original).Scan(&id)
	return id, err
}

func (r *PostgresURLRepository) FindByID(ctx context.Context, id int64) (*URL, error) {
	var u URL
	err := r.db.QueryRowContext(ctx, "SELECT original, clicks FROM urls WHERE id = $1", id).Scan(&u.Original, &u.Clicks)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresURLRepository) IncrementClicks(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE urls SET clicks = clicks + 1 WHERE id = $1", id)
	return err
}
