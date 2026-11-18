package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"url-shortener/internal/domain"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(ctx context.Context, addr, password string, ttl time.Duration) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Cache{client: rdb, ttl: ttl}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Set(ctx context.Context, shortCode, originalURL string) error {
	return c.client.Set(ctx, shortCode, originalURL, c.ttl).Err()
}

func (c *Cache) Get(ctx context.Context, shortCode string) (string, error) {
	val, err := c.client.Get(ctx, shortCode).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrURLNotFound
		}
		return "", err
	}
	return val, nil
}
