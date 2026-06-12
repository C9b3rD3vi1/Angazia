package models

import "time"

type Conversation struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Subject   string    `json:"subject" gorm:"size:255;not null;default:''"`
	JobID     *string   `json:"job_id,omitempty" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Participants []ConversationParticipant `json:"participants,omitempty" gorm:"foreignKey:ConversationID"`
	LastMessage  *Message                  `json:"last_message,omitempty" gorm:"foreignKey:ConversationID;references:ID"`
}

type ConversationParticipant struct {
	ID             string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID string    `json:"conversation_id" gorm:"type:uuid;not null;uniqueIndex:idx_conv_user;index"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_conv_user;index"`
	LastReadAt     time.Time `json:"last_read_at" gorm:"autoCreateTime"`
	IsArchived     bool      `json:"is_archived" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Message struct {
	ID             string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConversationID string    `json:"conversation_id" gorm:"type:uuid;not null;index"`
	SenderID       string    `json:"sender_id" gorm:"type:uuid;not null;index"`
	Content        string    `json:"content" gorm:"type:text;not null"`
	IsRead         bool      `json:"is_read" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;index"`

	Sender *User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
}

type ConversationListResponse struct {
	Conversations []*Conversation `json:"conversations"`
	Total         int64           `json:"total"`
	UnreadCount   int             `json:"unread_count"`
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
	TotalPages    int             `json:"total_pages"`
}

type MessageListResponse struct {
	Messages   []*Message `json:"messages"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}

func (Conversation) TableName() string                  { return "conversations" }
func (ConversationParticipant) TableName() string        { return "conversation_participants" }
func (Message) TableName() string                        { return "messages" }
