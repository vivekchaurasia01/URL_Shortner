package cache

import (
    "context"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
    return &RedisCache{
        client: redis.NewClient(&redis.Options{Addr: addr}),
    }
}

// Get returns the long URL for a short code, error on cache miss
func (c *RedisCache) Get(shortURL string) (string, error) {
    return c.client.Get(context.Background(), shortURL).Result()
}

// Set stores short -> long URL mapping
// No TTL — maxmemory-policy allkeys-lru handles eviction automatically
// LRU naturally keeps hot URLs in cache without us guessing a duration
func (c *RedisCache) Set(shortURL string, longURL string) error {
    return c.client.Set(context.Background(), shortURL, longURL, 0).Err()
}