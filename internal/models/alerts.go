package models

import (
	"time"
)

// AlertHistory tracks when alerts were sent
type AlertHistory struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SavedSearchID string     `json:"saved_search_id" gorm:"type:uuid;not null;index"`
	JobsFound     int        `json:"jobs_found" gorm:"default:0"`
	JobsSent      int        `json:"jobs_sent" gorm:"default:0"`
	JobIDs        []string   `json:"job_ids" gorm:"type:jsonb"`
	SentAt        time.Time  `json:"sent_at" gorm:"autoCreateTime;index"`
	
	// Relationships
	SavedSearch   *SavedSearch `json:"saved_search,omitempty" gorm:"foreignKey:SavedSearchID"`
}

// AlertSettings represents user's notification preferences
type AlertSettings struct {
	ID                string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	
	// Email settings
	EmailEnabled      bool      `json:"email_enabled" gorm:"default:true"`
	EmailDigestHour   int       `json:"email_digest_hour" gorm:"default:8"` // 8 AM
	EmailDigestDay    int       `json:"email_digest_day" gorm:"default:1"`  // Monday = 1
	
	// Push notifications
	PushEnabled       bool      `json:"push_enabled" gorm:"default:false"`
	
	// Alert types
	NewJobAlerts      bool      `json:"new_job_alerts" gorm:"default:true"`
	ApplicationUpdates bool     `json:"application_updates" gorm:"default:true"`
	MarketingEmails   bool      `json:"marketing_emails" gorm:"default:false"`
	
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User              *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (AlertHistory) TableName() string {
	return "alert_history"
}

func (AlertSettings) TableName() string {
	return "alert_settings"
}