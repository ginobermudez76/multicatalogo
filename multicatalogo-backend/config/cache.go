package config

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService define la abstracción de operaciones de caché requerida para Inyección de Dependencias.
type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RedisCacheService implementa la interfaz CacheService usando el cliente de Redis.
type RedisCacheService struct {
	client *redis.Client
}

// NewRedisCacheService crea una nueva instancia de RedisCacheService.
func NewRedisCacheService(client *redis.Client) *RedisCacheService {
	return &RedisCacheService{client: client}
}

func (r *RedisCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if r.client == nil {
		return nil, redis.Nil
	}
	return r.client.Get(ctx, key).Bytes()
}

func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if r.client == nil {
		return nil
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return nil
	}
	return r.client.Del(ctx, key).Err()
}
