package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/komoru/url-shortener/internal/cache"
	"github.com/komoru/url-shortener/internal/model"
	"github.com/komoru/url-shortener/internal/repository"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLen  = 7
)

type URLService struct {
	repo    repository.URLRepository
	cache   *cache.Cache
	baseURL string
}

func New(repo repository.URLRepository, c *cache.Cache, baseURL string) *URLService {
	return &URLService{repo: repo, cache: c, baseURL: baseURL}
}

func (s *URLService) Shorten(ctx context.Context, originalURL string) (*model.URL, error) {
	code, err := generateCode(codeLen)
	if err != nil {
		return nil, err
	}

	u := &model.URL{
		ShortCode:   code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	s.cache.Set(code, originalURL)
	return u, nil
}

// Resolve returns original URL by short code, incrementing click counter asynchronously.
func (s *URLService) Resolve(ctx context.Context, code string) (string, error) {
	if original, ok := s.cache.Get(code); ok {
		go s.repo.IncrementClicks(context.Background(), code) //nolint:errcheck
		return original, nil
	}

	u, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return "", err
	}

	s.cache.Set(code, u.OriginalURL)
	go s.repo.IncrementClicks(context.Background(), code) //nolint:errcheck
	return u.OriginalURL, nil
}

func (s *URLService) Stats(ctx context.Context, code string) (*model.URL, error) {
	return s.repo.GetByShortCode(ctx, code)
}

func (s *URLService) ShortURL(code string) string {
	return s.baseURL + "/" + code
}

func generateCode(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	alphabetLen := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[idx.Int64()])
	}
	return sb.String(), nil
}
