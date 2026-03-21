package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/komoru/url-shortener/internal/cache"
	"github.com/komoru/url-shortener/internal/model"
	"github.com/komoru/url-shortener/internal/repository"
	"github.com/komoru/url-shortener/internal/service"
)

type mockRepo struct {
	urls map[string]*model.URL
}

func newMockRepo() *mockRepo {
	return &mockRepo{urls: make(map[string]*model.URL)}
}

func (m *mockRepo) Create(_ context.Context, u *model.URL) error {
	u.ID = int64(len(m.urls) + 1)
	m.urls[u.ShortCode] = u
	return nil
}

func (m *mockRepo) GetByShortCode(_ context.Context, code string) (*model.URL, error) {
	u, ok := m.urls[code]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) IncrementClicks(_ context.Context, code string) error {
	if u, ok := m.urls[code]; ok {
		u.Clicks++
	}
	return nil
}

func newTestService(t *testing.T) (*service.URLService, *mockRepo) {
	t.Helper()
	repo := newMockRepo()
	c := cache.New(5 * time.Minute)
	t.Cleanup(c.Stop)
	svc := service.New(repo, c, "http://localhost:8080")
	return svc, repo
}

func TestShorten_CreatesURL(t *testing.T) {
	svc, repo := newTestService(t)

	u, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if len(u.ShortCode) != 7 {
		t.Errorf("ожидали длину кода 7, получили %d", len(u.ShortCode))
	}
	if _, ok := repo.urls[u.ShortCode]; !ok {
		t.Error("URL не сохранён в репозитории")
	}
}

func TestShorten_ShortURLFormat(t *testing.T) {
	svc, _ := newTestService(t)

	u, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	want := "http://localhost:8080/" + u.ShortCode
	if got := svc.ShortURL(u.ShortCode); got != want {
		t.Errorf("ShortURL() = %q, want %q", got, want)
	}
}

func TestResolve_FromCache(t *testing.T) {
	svc, _ := newTestService(t)

	original := "https://example.com/long-path"
	u, _ := svc.Shorten(context.Background(), original)

	got, err := svc.Resolve(context.Background(), u.ShortCode)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != original {
		t.Errorf("Resolve() = %q, want %q", got, original)
	}
}

func TestResolve_FromRepository(t *testing.T) {
	repo := newMockRepo()
	c := cache.New(5 * time.Minute)
	t.Cleanup(c.Stop)
	svc := service.New(repo, c, "http://localhost:8080")

	_ = repo.Create(context.Background(), &model.URL{
		ShortCode:   "abc1234",
		OriginalURL: "https://avito.ru",
		CreatedAt:   time.Now(),
	})

	got, err := svc.Resolve(context.Background(), "abc1234")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got != "https://avito.ru" {
		t.Errorf("Resolve() = %q, want https://avito.ru", got)
	}
}

func TestResolve_NotFound(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Resolve(context.Background(), "notexist")
	if err != repository.ErrNotFound {
		t.Errorf("Resolve() error = %v, want ErrNotFound", err)
	}
}

func TestStats_ReturnsURL(t *testing.T) {
	svc, _ := newTestService(t)

	original := "https://avito.ru"
	u, _ := svc.Shorten(context.Background(), original)

	stats, err := svc.Stats(context.Background(), u.ShortCode)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.OriginalURL != original {
		t.Errorf("Stats().OriginalURL = %q, want %q", stats.OriginalURL, original)
	}
}
