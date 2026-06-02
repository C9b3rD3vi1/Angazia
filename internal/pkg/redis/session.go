package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type SessionData struct {
	UserID    string                 `json:"user_id"`
	Role      string                 `json:"role"`
	Email     string                 `json:"email"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

type SessionManager struct {
	client *RedisClient
	ttl    time.Duration
	prefix string
}

func NewSessionManager(client *RedisClient, ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &SessionManager{
		client: client,
		ttl:    ttl,
		prefix: "session:",
	}
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *SessionManager) CreateSession(ctx context.Context, userID, role, email string, metadata map[string]interface{}) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	data := SessionData{
		UserID:    userID,
		Role:      role,
		Email:     email,
		Metadata:  metadata,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}

	key := s.prefix + sessionID
	if err := s.client.Set(ctx, key, data, s.ttl); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return sessionID, nil
}

func (s *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := s.prefix + sessionID
	var data SessionData
	if err := s.client.Get(ctx, key, &data); err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	if time.Now().After(data.ExpiresAt) {
		s.client.Delete(ctx, key)
		return nil, fmt.Errorf("session expired")
	}

	return &data, nil
}

func (s *SessionManager) UpdateSession(ctx context.Context, sessionID string, metadata map[string]interface{}) error {
	data, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if data.Metadata == nil {
		data.Metadata = make(map[string]interface{})
	}
	for k, v := range metadata {
		data.Metadata[k] = v
	}

	remaining := time.Until(data.ExpiresAt)
	if remaining <= 0 {
		return fmt.Errorf("session expired")
	}

	key := s.prefix + sessionID
	return s.client.Set(ctx, key, data, remaining)
}

func (s *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := s.prefix + sessionID
	return s.client.Delete(ctx, key)
}

func (s *SessionManager) DeleteUserSessions(ctx context.Context, userID string) error {
	pattern := s.prefix + "*"
	keys, err := s.client.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		var data SessionData
		if err := s.client.Get(ctx, key, &data); err != nil {
			continue
		}
		if data.UserID == userID {
			s.client.Delete(ctx, key)
		}
	}

	return nil
}

func (s *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	data, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	data.ExpiresAt = time.Now().Add(s.ttl)
	key := s.prefix + sessionID
	return s.client.Set(ctx, key, data, s.ttl)
}

func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}
