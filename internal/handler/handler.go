package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/komoru/url-shortener/internal/repository"
	"github.com/komoru/url-shortener/internal/service"
)

type Handler struct {
	svc *service.URLService
}

func New(svc *service.URLService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Post("/api/shorten", h.Shorten)
	r.Get("/api/stats/{code}", h.Stats)
	r.Get("/{code}", h.Redirect)
	r.Handle("/*", http.FileServer(http.Dir("./frontend")))

	return r
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeError(w, "url is required", http.StatusBadRequest)
		return
	}

	u, err := h.svc.Shorten(r.Context(), req.URL)
	if err != nil {
		writeError(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		ShortURL:  h.svc.ShortURL(u.ShortCode),
		ShortCode: u.ShortCode,
	})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	original, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, original, http.StatusMovedPermanently)
}

type statsResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int64  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
	ShortURL    string `json:"short_url"`
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	u, err := h.svc.Stats(r.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, "not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, statsResponse{
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalURL,
		Clicks:      u.Clicks,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		ShortURL:    h.svc.ShortURL(u.ShortCode),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
