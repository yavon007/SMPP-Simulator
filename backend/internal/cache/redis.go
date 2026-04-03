package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides Redis caching capabilities
type RedisCache struct {
	client *redis.Client
	enabled bool
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(host, port, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis connected: %s:%s", host, port)
	return &RedisCache{
		client:  client,
		enabled: true,
	}, nil
}

// NewNoopCache creates a no-op cache (for when Redis is not available)
func NewNoopCache() *RedisCache {
	return &RedisCache{
		client:  nil,
		enabled: false,
	}
}

// IsEnabled returns whether Redis is enabled
func (c *RedisCache) IsEnabled() bool {
	return c.enabled
}

// Set stores a key-value pair with expiration
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value by key
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.enabled {
		return fmt.Errorf("cache not enabled")
	}

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.enabled {
		return nil
	}

	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if !c.enabled {
		return false, nil
	}

	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Increment increments a counter
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	if !c.enabled {
		return 0, nil
	}

	return c.client.Incr(ctx, key).Result()
}

// IncrementWithExpire increments a counter and sets expiration
func (c *RedisCache) IncrementWithExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if !c.enabled {
		return 0, nil
	}

	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return result, err
	}

	// Set expiration only on first increment
	if result == 1 {
		c.client.Expire(ctx, key, expiration)
	}

	return result, nil
}

// SetNX sets a key-value pair only if the key doesn't exist
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if !c.enabled {
		return true, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.SetNX(ctx, key, data, expiration).Result()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// SessionCache manages session caching
type SessionCache struct {
	cache *RedisCache
}

// NewSessionCache creates a new session cache
func NewSessionCache(cache *RedisCache) *SessionCache {
	return &SessionCache{cache: cache}
}

// CacheSession stores session info in cache
func (sc *SessionCache) CacheSession(ctx context.Context, sessionID string, sessionData interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sc.cache.Set(ctx, key, sessionData, 30*time.Minute)
}

// GetSession retrieves cached session info
func (sc *SessionCache) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sc.cache.Get(ctx, key, dest)
}

// DeleteSession removes cached session info
func (sc *SessionCache) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sc.cache.Delete(ctx, key)
}

// RateLimitCache manages rate limiting with Redis
type RateLimitCache struct {
	cache *RedisCache
}

// NewRateLimitCache creates a new rate limit cache
func NewRateLimitCache(cache *RedisCache) *RateLimitCache {
	return &RateLimitCache{cache: cache}
}

// CheckRateLimit checks if a request is within rate limit
func (rc *RateLimitCache) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	if !rc.cache.IsEnabled() {
		// If Redis is not available, allow the request
		return true, 0, nil
	}

	fullKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := rc.cache.IncrementWithExpire(ctx, fullKey, window)
	if err != nil {
		return true, 0, err
	}

	return count <= int64(limit), int(count), nil
}
