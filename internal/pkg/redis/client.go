package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
)

type RedisClient struct {
	client *redis.Client
	cfg    *config.Config
}

func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully")

	return &RedisClient{
		client: client,
		cfg:    cfg,
	}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Cache operations

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// Hash operations (for storing objects by field)

func (r *RedisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, key, field, data).Err()
}

func (r *RedisClient) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	data, err := r.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// Set operations (for unique items)

func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// Sorted Set operations (for rankings/leaderboards)

func (r *RedisClient) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (r *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(ctx, key, start, stop).Result()
}

func (r *RedisClient) ZIncrBy(ctx context.Context, key string, increment float64, member string) error {
	return r.client.ZIncrBy(ctx, key, increment, member).Err()
}

func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, min, max float64) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", min),
		Max: fmt.Sprintf("%f", max),
	}).Result()
}

func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Queue operations (for background jobs)

func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *RedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return r.client.BRPop(ctx, timeout, keys...).Result()
}

// Rate limiting operations

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Cached function wrapper

func (r *RedisClient) CacheGetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), dest interface{}) error {
	// Check cache
	err := r.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	// Execute function
	data, err := fn()
	if err != nil {
		return err
	}

	// Store in cache
	if err := r.Set(ctx, key, data, ttl); err != nil {
		log.Printf("Failed to cache data: %v", err)
	}

	// Copy to destination
	jsonData, _ := json.Marshal(data)
	return json.Unmarshal(jsonData, dest)
}
