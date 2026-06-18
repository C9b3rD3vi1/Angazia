package models

import "time"

type ContactSubmission struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:255;not null"`
	Subject   string    `json:"subject" gorm:"size:255"`
	Message   string    `json:"message" gorm:"type:text;not null"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ContactSubmission) TableName() string {
	return "contact_submissions"
}

type ContactSubmissionListResponse struct {
	Submissions []*ContactSubmission `json:"submissions"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	TotalPages  int                  `json:"total_pages"`
	UnreadCount int64                `json:"unread_count"`
}
