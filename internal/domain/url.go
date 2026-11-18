package domain

import (
	"errors"
	"time"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("short code already exists")
	ErrInvalidURL  = errors.New("invalid url format")
)

type URL struct {
	ID          int64
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}
