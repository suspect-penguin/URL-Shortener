package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/komoru/url-shortener/internal/cache"
	"github.com/komoru/url-shortener/internal/handler"
	"github.com/komoru/url-shortener/internal/model"
	"github.com/komoru/url-shortener/internal/repository"
	"github.com/komoru/url-shortener/internal/service"
)

type mockRepo struct {
	urls map[string]*model.URL
}

func newMockRepo() *mockRepo { return &mockRepo{urls: make(map[string]*model.URL)} }

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

func newTestRouter(t *testing.T) (http.Handler, *mockRepo) {
	t.Helper()
	repo := newMockRepo()
	c := cache.New(5 * time.Minute)
	t.Cleanup(c.Stop)
	svc := service.New(repo, c, "http://localhost:8080")
	return handler.New(svc).Routes(), repo
}

func TestShortenHandler(t *testing.T) {
	router, _ := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("ожидали 201, получили %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["short_code"] == "" {
		t.Error("ожидали short_code в ответе")
	}
	if resp["short_url"] == "" {
		t.Error("ожидали short_url в ответе")
	}
}

func TestShortenHandler_EmptyURL(t *testing.T) {
	router, _ := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{"url": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ожидали 400, получили %d", w.Code)
	}
}

func TestStatsHandler(t *testing.T) {
	router, repo := newTestRouter(t)

	_ = repo.Create(context.Background(), &model.URL{
		ShortCode:   "abc1234",
		OriginalURL: "https://avito.ru",
		CreatedAt:   time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/abc1234", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["original_url"] != "https://avito.ru" {
		t.Errorf("неожиданный original_url: %v", resp["original_url"])
	}
}

func TestStatsHandler_NotFound(t *testing.T) {
	router, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/stats/notexist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("ожидали 404, получили %d", w.Code)
	}
}
