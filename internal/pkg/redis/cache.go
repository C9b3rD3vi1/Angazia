package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type CacheOptions struct {
	TTL        time.Duration
	Tags       []string
	Compress   bool
	SkipEmpty  bool
}

type CacheService struct {
	client *RedisClient
	prefix string
}

func NewCacheService(client *RedisClient) *CacheService {
	return &CacheService{
		client: client,
		prefix: "cache:",
	}
}

func (s *CacheService) generateKey(parts ...string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s", parts)))
	return s.prefix + hex.EncodeToString(hash[:])
}

func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	return s.client.Get(ctx, key, dest)
}

func (s *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl)
}

func (s *CacheService) Delete(ctx context.Context, keys ...string) error {
	return s.client.Delete(ctx, keys...)
}

func (s *CacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), dest interface{}) error {
	err := s.client.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	data, err := fn()
	if err != nil {
		return err
	}

	if err := s.client.Set(ctx, key, data, ttl); err != nil {
		log.Printf("Failed to cache data for key %s: %v", key, err)
	}

	jsonData, _ := json.Marshal(data)
	return json.Unmarshal(jsonData, dest)
}

func (s *CacheService) InvalidateByPrefix(ctx context.Context, prefix string) error {
	pattern := s.prefix + prefix + "*"
	keys, err := s.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.client.Delete(ctx, keys...)
	}
	return nil
}

func (s *CacheService) InvalidateAll(ctx context.Context) error {
	pattern := s.prefix + "*"
	keys, err := s.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.client.Delete(ctx, keys...)
	}
	return nil
}

func (s *CacheService) WarmCache(ctx context.Context, entries []CacheWarmEntry) {
	for _, entry := range entries {
		var dest interface{}
		err := s.GetOrSet(ctx, entry.Key, entry.TTL, entry.Fn, &dest)
		if err != nil {
			log.Printf("Failed to warm cache for key %s: %v", entry.Key, err)
		}
	}
}

type CacheWarmEntry struct {
	Key string
	TTL time.Duration
	Fn  func() (interface{}, error)
}

func (s *CacheService) BuildListCacheKey(entityType string, filters map[string]interface{}, page, limit int) string {
	filterBytes, _ := json.Marshal(filters)
	raw := fmt.Sprintf("%s:%s:%d:%d", entityType, string(filterBytes), page, limit)
	hash := sha256.Sum256([]byte(raw))
	return s.prefix + "list:" + hex.EncodeToString(hash[:])
}

func (s *CacheService) BuildDetailCacheKey(entityType, id string) string {
	raw := fmt.Sprintf("%s:%s", entityType, id)
	hash := sha256.Sum256([]byte(raw))
	return s.prefix + "detail:" + hex.EncodeToString(hash[:])
}
