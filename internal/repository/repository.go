package repository

import (
	"context"
	"errors"

	"github.com/komoru/url-shortener/internal/model"
)

var ErrNotFound = errors.New("url not found")

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	IncrementClicks(ctx context.Context, shortCode string) error
}
