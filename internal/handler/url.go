package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"url-shortener/internal/domain"

	"github.com/gin-gonic/gin"
)

type shortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	code, err := h.service.Shorten(c.Request.Context(), req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.log.Error("shorten failed", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := shortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", strings.TrimRight(h.baseURL, "/"), code),
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	originalURL, err := h.service.Resolve(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
			return
		}
		h.log.Error("resolve failed", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}
