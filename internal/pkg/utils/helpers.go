package utils

import (
	"os"
	
    "math/rand"
    "time"
)

func IsDevelopment() bool {
	return os.Getenv("ENVIRONMENT") == "development"
}

func IsProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

// GenerateRandomString creates a random string of the specified length
func GenerateRandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
    
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[seededRand.Intn(len(charset))]
    }
    return string(b)
}