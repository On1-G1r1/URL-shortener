package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPPort    string        `env:"HTTP_PORT" env-default:"8080"`
	BaseURL     string        `env:"BASE_URL" env-default:"http://localhost:8080"`
	PostgresDSN string        `env:"POSTGRES_DSN" env-required:"true"`
	RedisAddr   string        `env:"REDIS_ADDR" env-required:"true"`
	RedisPass   string        `env:"REDIS_PASSWORD" env-default:""`
	RedisTTL    time.Duration `env:"REDIS_TTL" env-default:"24h"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg
}
