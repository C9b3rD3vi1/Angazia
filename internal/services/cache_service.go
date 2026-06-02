package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/redis"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type CacheService interface {
	// Job caching
	CacheJob(ctx context.Context, job *models.Job) error
	GetCachedJob(ctx context.Context, jobID string) (*models.Job, error)
	InvalidateJob(ctx context.Context, jobID string) error
	
	// Job list caching
	CacheJobList(ctx context.Context, filters map[string]interface{}, page, limit int, jobs []*models.Job, total int64) error
	GetCachedJobList(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Job, int64, error)
	
	// User caching
	CacheUser(ctx context.Context, user *models.User) error
	GetCachedUser(ctx context.Context, userID string) (*models.User, error)
	InvalidateUser(ctx context.Context, userID string) error
	
	// Application statistics
	CacheApplicationStats(ctx context.Context, userID string, stats interface{}) error
	GetCachedApplicationStats(ctx context.Context, userID string, dest interface{}) error
	
	// Popular searches
	RecordPopularSearch(ctx context.Context, query string) error
	GetPopularSearches(ctx context.Context, limit int) ([]string, error)
	
	// Rate limiting
	IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type CacheServiceImpl struct {
	redisClient *redis.RedisClient
	jobRepo     repository.JobRepository
	userRepo    repository.UserRepository
}

func NewCacheService(
	redisClient *redis.RedisClient,
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
) CacheService {
	return &CacheServiceImpl{
		redisClient: redisClient,
		jobRepo:     jobRepo,
		userRepo:    userRepo,
	}
}

func (s *CacheServiceImpl) CacheJob(ctx context.Context, job *models.Job) error {
	key := fmt.Sprintf("job:%s", job.ID)
	return s.redisClient.Set(ctx, key, job, 1*time.Hour)
}

func (s *CacheServiceImpl) GetCachedJob(ctx context.Context, jobID string) (*models.Job, error) {
	key := fmt.Sprintf("job:%s", jobID)
	var job models.Job
	err := s.redisClient.Get(ctx, key, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *CacheServiceImpl) InvalidateJob(ctx context.Context, jobID string) error {
	key := fmt.Sprintf("job:%s", jobID)
	return s.redisClient.Delete(ctx, key)
}

func (s *CacheServiceImpl) CacheJobList(ctx context.Context, filters map[string]interface{}, page, limit int, jobs []*models.Job, total int64) error {
	// Generate cache key from filters
	key := s.generateListCacheKey("jobs", filters, page, limit)
	
	cacheData := map[string]interface{}{
		"jobs":  jobs,
		"total": total,
	}
	
	return s.redisClient.Set(ctx, key, cacheData, 5*time.Minute)
}

func (s *CacheServiceImpl) GetCachedJobList(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Job, int64, error) {
	key := s.generateListCacheKey("jobs", filters, page, limit)
	
	var cacheData struct {
		Jobs  []*models.Job `json:"jobs"`
		Total int64         `json:"total"`
	}
	
	err := s.redisClient.Get(ctx, key, &cacheData)
	if err != nil {
		return nil, 0, err
	}
	
	return cacheData.Jobs, cacheData.Total, nil
}

func (s *CacheServiceImpl) CacheUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:%s", user.ID)
	return s.redisClient.Set(ctx, key, user, 30*time.Minute)
}

func (s *CacheServiceImpl) GetCachedUser(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("user:%s", userID)
	var user models.User
	err := s.redisClient.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *CacheServiceImpl) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	return s.redisClient.Delete(ctx, key)
}

func (s *CacheServiceImpl) CacheApplicationStats(ctx context.Context, userID string, stats interface{}) error {
	key := fmt.Sprintf("stats:applications:%s", userID)
	return s.redisClient.Set(ctx, key, stats, 15*time.Minute)
}

func (s *CacheServiceImpl) GetCachedApplicationStats(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("stats:applications:%s", userID)
	return s.redisClient.Get(ctx, key, dest)
}

func (s *CacheServiceImpl) RecordPopularSearch(ctx context.Context, query string) error {
	key := "popular_searches"
	// Increment score for this search term
	return s.redisClient.ZIncrBy(ctx, key, 1, query)
}

func (s *CacheServiceImpl) GetPopularSearches(ctx context.Context, limit int) ([]string, error) {
	key := "popular_searches"
	return s.redisClient.ZRevRange(ctx, key, 0, int64(limit-1))
}

func (s *CacheServiceImpl) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	rateKey := fmt.Sprintf("rate_limit:%s", key)
	
	count, err := s.redisClient.Incr(ctx, rateKey)
	if err != nil {
		return false, err
	}
	
	if count == 1 {
		s.redisClient.Expire(ctx, rateKey, window)
	}
	
	return count > int64(limit), nil
}

func (s *CacheServiceImpl) generateListCacheKey(prefix string, filters map[string]interface{}, page, limit int) string {
	filterBytes, _ := json.Marshal(filters)
	raw := fmt.Sprintf("%s:p%d:l%d:%s", prefix, page, limit, string(filterBytes))
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%s:list:%s", prefix, hex.EncodeToString(hash[:]))
}