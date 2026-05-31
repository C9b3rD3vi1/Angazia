package models

import (
	"time"
)

// Notification represents a user notification
type Notification struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string     `json:"user_id" gorm:"type:uuid;not null;index"`
	Type        string     `json:"type" gorm:"size:50;not null;index"` // application, interview, message, alert, system
	Title       string     `json:"title" gorm:"size:255;not null"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	Metadata    JSONMap    `json:"metadata" gorm:"type:jsonb"`
	ActionURL   string     `json:"action_url" gorm:"size:512"`
	Icon        string     `json:"icon" gorm:"size:50"`
	Priority    string     `json:"priority" gorm:"size:20;default:'normal'"` // high, normal, low
	IsRead      bool       `json:"is_read" gorm:"default:false;index"`
	IsArchived  bool       `json:"is_archived" gorm:"default:false"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	ID                string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	
	// Push notifications
	PushEnabled       bool      `json:"push_enabled" gorm:"default:true"`
	PushSound         bool      `json:"push_sound" gorm:"default:true"`
	
	// Email notifications
	EmailEnabled      bool      `json:"email_enabled" gorm:"default:true"`
	EmailDigest       bool      `json:"email_digest" gorm:"default:false"`
	
	// In-app notifications
	InAppEnabled      bool      `json:"in_app_enabled" gorm:"default:true"`
	
	// Notification types
	ApplicationUpdates bool     `json:"application_updates" gorm:"default:true"`
	JobAlerts         bool      `json:"job_alerts" gorm:"default:true"`
	InterviewReminders bool     `json:"interview_reminders" gorm:"default:true"`
	Messages          bool      `json:"messages" gorm:"default:true"`
	SystemAlerts      bool      `json:"system_alerts" gorm:"default:true"`
	Marketing         bool      `json:"marketing" gorm:"default:false"`
	
	// Quiet hours
	QuietHoursEnabled bool      `json:"quiet_hours_enabled" gorm:"default:false"`
	QuietStartHour    int       `json:"quiet_start_hour" gorm:"default:22"` // 10 PM
	QuietEndHour      int       `json:"quiet_end_hour" gorm:"default:8"`    // 8 AM
	QuietTimezone     string    `json:"quiet_timezone" gorm:"default:'Africa/Nairobi'"`
	
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User              *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// NotificationCounts represents unread notification counts
type NotificationCounts struct {
	TotalUnread       int            `json:"total_unread"`
	ByType            map[string]int `json:"by_type"`
	HighPriorityCount int            `json:"high_priority_count"`
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`      // notification, ping, pong, auth_required, auth_success
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NotificationListResponse represents paginated notifications
type NotificationListResponse struct {
	Notifications []*Notification `json:"notifications"`
	Total         int64           `json:"total"`
	UnreadCount   int             `json:"unread_count"`
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
	TotalPages    int             `json:"total_pages"`
}

// MarkReadRequest represents request to mark notifications as read
type MarkReadRequest struct {
	NotificationIDs []string `json:"notification_ids"`
	MarkAll         bool     `json:"mark_all"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (NotificationPreferences) TableName() string {
	return "notification_preferences"
}