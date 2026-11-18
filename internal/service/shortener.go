package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/url"

	"url-shortener/internal/domain"
)

const (
	alphabet   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	codeLength = 6
	maxRetries = 5
)

type URLStorage interface {
	SaveURL(ctx context.Context, shortCode, originalURL string) error
	GetURL(ctx context.Context, shortCode string) (string, error)
}

type URLCache interface {
	Set(ctx context.Context, shortCode, originalURL string) error
	Get(ctx context.Context, shortCode string) (string, error)
}

type Shortener struct {
	storage URLStorage
	cache   URLCache
	log     *slog.Logger
}

func NewShortener(storage URLStorage, cache URLCache, log *slog.Logger) *Shortener {
	return &Shortener{
		storage: storage,
		cache:   cache,
		log:     log,
	}
}

func (s *Shortener) Shorten(ctx context.Context, rawURL string) (string, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", domain.ErrInvalidURL
	}

	for i := 0; i < maxRetries; i++ {
		code, err := generateBase62(codeLength)
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}

		err = s.storage.SaveURL(ctx, code, rawURL)
		if err == nil {
			// Опционально прогреваем кэш, игнорируя ошибку Redis
			if s.cache != nil {
				_ = s.cache.Set(ctx, code, rawURL)
			}
			return code, nil
		}
		if !errors.Is(err, domain.ErrURLExists) {
			return "", err
		}
	}

	return "", errors.New("failed to generate unique code after retries")
}

func (s *Shortener) Resolve(ctx context.Context, shortCode string) (string, error) {
	// 1. Проверяем кэш (Cache-Aside)
	if s.cache != nil {
		originalURL, err := s.cache.Get(ctx, shortCode)
		if err == nil {
			return originalURL, nil
		}
		if !errors.Is(err, domain.ErrURLNotFound) {
			// Отказоустойчивость: логируем падение Redis, но идем дальше в БД
			s.log.Warn("redis get failed, falling back to db", slog.String("err", err.Error()))
		}
	}

	// 2. Идем в PostgreSQL
	originalURL, err := s.storage.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// 3. Сохраняем в Redis для будущих запросов
	if s.cache != nil {
		if err := s.cache.Set(ctx, shortCode, originalURL); err != nil {
			s.log.Warn("redis set failed", slog.String("err", err.Error()))
		}
	}

	return originalURL, nil
}

func generateBase62(length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[num.Int64()]
	}
	return string(result), nil
}
