package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type tokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *tokenRepository {
	return &tokenRepository{client: client}
}

func (r *tokenRepository) StoreToken(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
