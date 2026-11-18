package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"url-shortener/internal/domain"
)

type mockStorage struct {
	data map[string]string
}

func (m *mockStorage) SaveURL(_ context.Context, code, url string) error {
	if _, exists := m.data[code]; exists {
		return domain.ErrURLExists
	}
	m.data[code] = url
	return nil
}

func (m *mockStorage) GetURL(_ context.Context, code string) (string, error) {
	url, ok := m.data[code]
	if !ok {
		return "", domain.ErrURLNotFound
	}
	return url, nil
}

type mockCache struct {
	data     map[string]string
	failNext bool
}

func (m *mockCache) Set(_ context.Context, code, url string) error {
	if m.failNext {
		return errors.New("redis down")
	}
	m.data[code] = url
	return nil
}

func (m *mockCache) Get(_ context.Context, code string) (string, error) {
	if m.failNext {
		return "", errors.New("redis down")
	}
	url, ok := m.data[code]
	if !ok {
		return "", domain.ErrURLNotFound
	}
	return url, nil
}

func TestShortener_Flow(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	store := &mockStorage{data: make(map[string]string)}
	cache := &mockCache{data: make(map[string]string)}
	svc := NewShortener(store, cache, logger)

	ctx := context.Background()
	target := "https://example.com"

	// Тест создания
	code, err := svc.Shorten(ctx, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != codeLength {
		t.Fatalf("expected code length %d, got %d", codeLength, len(code))
	}

	// Тест получения при упавшем Redis (отказоустойчивость)
	cache.failNext = true
	resolved, err := svc.Resolve(ctx, code)
	if err != nil {
		t.Fatalf("expected fallback to db, got error: %v", err)
	}
	if resolved != target {
		t.Fatalf("expected %s, got %s", target, resolved)
	}

	// Тест невалидного URL
	if _, err := svc.Shorten(ctx, "not-a-url"); !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}
