package utils

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/Angazia/internal/config"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var cfg *config.Config

func InitJWT(config *config.Config) {
	cfg = config
}

func GenerateJWT(userID string, role string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWTExpiryHours) * time.Hour)
	
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.AppName,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}

func GetUserIDFromContext(c *fiber.Ctx) string {
	userID := c.Locals("user_id")
	if userID == nil {
		return ""
	}
	return userID.(string)
}

func GetUserRoleFromContext(c *fiber.Ctx) string {
	role := c.Locals("user_role")
	if role == nil {
		return ""
	}
	return role.(string)
}

func HashPassword(password string) (string, error) {
	// Implement bcrypt hashing
	// This is a placeholder
	return password, nil
}