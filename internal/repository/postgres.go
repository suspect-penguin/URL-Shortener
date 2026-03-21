package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/komoru/url-shortener/internal/model"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgres(db *pgxpool.Pool) URLRepository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, url *model.URL) error {
	const q = `
		INSERT INTO urls (short_code, original_url, created_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	return r.db.QueryRow(ctx, q, url.ShortCode, url.OriginalURL, url.CreatedAt).Scan(&url.ID)
}

func (r *postgresRepo) GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	const q = `SELECT id, short_code, original_url, clicks, created_at FROM urls WHERE short_code = $1`
	u := &model.URL{}
	err := r.db.QueryRow(ctx, q, shortCode).Scan(
		&u.ID, &u.ShortCode, &u.OriginalURL, &u.Clicks, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *postgresRepo) IncrementClicks(ctx context.Context, shortCode string) error {
	_, err := r.db.Exec(ctx, `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`, shortCode)
	return err
}
