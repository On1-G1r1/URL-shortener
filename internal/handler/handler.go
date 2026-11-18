package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery(), h.loggingMiddleware())

	router.POST("/shorten", h.shorten)
	router.GET("/:code", h.redirect)

	return router
}

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		h.log.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
