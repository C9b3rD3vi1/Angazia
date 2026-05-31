// internal/services/token_service.go
package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
)

type TokenService interface {
	GenerateUnsubscribeToken(email, userID string) (string, error)
	ValidateUnsubscribeToken(token string) (bool, error)
	GenerateEmailVerificationToken(email, userID string) (string, error)
	GeneratePasswordResetToken(email, userID string) (string, error)
}

type TokenServiceImpl struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) TokenService {
	return &TokenServiceImpl{
		cfg: cfg,
	}
}

func (s *TokenServiceImpl) GenerateUnsubscribeToken(email, userID string) (string, error) {
	// Create a unique token using multiple entropy sources
	timestamp := time.Now().UnixNano()
	uuid := uuid.New().String()
	
	// Combine inputs for hashing
	data := fmt.Sprintf("%s:%s:%d:%s", email, userID, timestamp, uuid)
	
	// Create HMAC-like hash with secret
	hash := sha256.New()
	hash.Write([]byte(data))
	hash.Write([]byte(s.cfg.JWTSecret))
	hashBytes := hash.Sum(nil)
	
	// Encode to URL-safe base64
	token := base64.URLEncoding.EncodeToString(hashBytes)
	
	// Add random suffix for additional security
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	randomSuffix := base64.URLEncoding.EncodeToString(randomBytes)[:8]
	
	// Final token: hash + random suffix
	finalToken := token[:32] + randomSuffix
	
	return finalToken, nil
}

func (s *TokenServiceImpl) ValidateUnsubscribeToken(token string) (bool, error) {
	// Token validation is done in repository layer
	// This method is for additional validation if needed
	if len(token) < 40 {
		return false, fmt.Errorf("invalid token length")
	}
	return true, nil
}

func (s *TokenServiceImpl) GenerateEmailVerificationToken(email, userID string) (string, error) {
	timestamp := time.Now().UnixNano()
	uuid := uuid.New().String()
	
	data := fmt.Sprintf("verify:%s:%s:%d:%s", email, userID, timestamp, uuid)
	
	hash := sha256.New()
	hash.Write([]byte(data))
	hash.Write([]byte(s.cfg.JWTSecret))
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *TokenServiceImpl) GeneratePasswordResetToken(email, userID string) (string, error) {
	timestamp := time.Now().UnixNano()
	uuid := uuid.New().String()
	
	data := fmt.Sprintf("reset:%s:%s:%d:%s", email, userID, timestamp, uuid)
	
	hash := sha256.New()
	hash.Write([]byte(data))
	hash.Write([]byte(s.cfg.JWTSecret))
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}