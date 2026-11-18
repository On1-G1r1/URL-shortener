package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type ShortenerService interface {
	Shorten(ctx context.Context, rawURL string) (string, error)
	Resolve(ctx context.Context, shortCode string) (string, error)
}

type Handler struct {
	service ShortenerService
	baseURL string
	log     *slog.Logger
}

func New(service ShortenerService, baseURL string, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		baseURL: baseURL,
		log:     log,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.shorten)
	mux.HandleFunc("GET /{code}", h.redirect)

	return h.loggingMiddleware(mux)
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		h.log.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
