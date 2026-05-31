// internal/models/unsubscribe.go
package models

import (
	"time"
)

// UnsubscribeToken represents a user's unsubscribe token for email notifications
type UnsubscribeToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null;size:255"`
	UserID    *string   `json:"user_id,omitempty" gorm:"type:uuid;index"`
	IsActive  bool      `json:"is_active" gorm:"default:true;index"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
}

func (UnsubscribeToken) TableName() string {
	return "unsubscribe_tokens"
}

// EmailPreferences stores user email notification preferences
type EmailPreferences struct {
	ID                    string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID                string    `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	
	// Notification types
	JobAlerts             bool      `json:"job_alerts" gorm:"default:true"`
	ApplicationUpdates    bool      `json:"application_updates" gorm:"default:true"`
	MarketingEmails       bool      `json:"marketing_emails" gorm:"default:true"`
	SecurityAlerts        bool      `json:"security_alerts" gorm:"default:true"`
	Newsletter            bool      `json:"newsletter" gorm:"default:false"`
	
	// Digest settings
	DigestFrequency       string    `json:"digest_frequency" gorm:"default:'daily';size:20"` // daily, weekly, never
	LastDigestSentAt      *time.Time `json:"last_digest_sent_at,omitempty"`
	
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EmailPreferences) TableName() string {
	return "email_preferences"
}