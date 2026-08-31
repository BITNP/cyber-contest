package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"national-defense-knowledge-quiz/internal/config"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.RedisAddr, err)
	}

	return rdb, nil
}

type KeyBuilder struct {
	prefix string
}

func NewKeyBuilder(prefix string) *KeyBuilder {
	return &KeyBuilder{prefix: prefix}
}

func (k *KeyBuilder) Key(parts ...string) string {
	all := make([]string, 0, len(parts)+1)
	all = append(all, k.prefix)
	all = append(all, parts...)
	return strings.Join(all, ":")
}
