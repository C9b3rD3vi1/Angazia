package models

import "time"

type Testimonial struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:uuid;not null;index"`
	UserName    string    `json:"user_name" gorm:"size:255;not null"`
	UserTitle   string    `json:"user_title" gorm:"size:255"`
	CompanyName string    `json:"company_name" gorm:"size:255"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	Rating      int       `json:"rating" gorm:"default:0"`
	IsApproved  bool      `json:"is_approved" gorm:"default:false"`
	IsFeatured  bool      `json:"is_featured" gorm:"default:false"`
	Role        string    `json:"role" gorm:"size:20;default:'employee';index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Testimonial) TableName() string {
	return "testimonials"
}

type ListTestimonialsParams struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Status     string `json:"status"`
	Role       string `json:"role"`
	IsFeatured *bool  `json:"is_featured"`
	Search     string `json:"search"`
}

type TestimonialListResponse struct {
	Testimonials []*Testimonial `json:"testimonials"`
	Total        int64          `json:"total"`
	Page         int            `json:"page"`
	Limit        int            `json:"limit"`
	TotalPages   int            `json:"total_pages"`
}

type CreateTestimonialRequest struct {
	Content     string `json:"content" validate:"required,min=20,max=1000"`
	UserTitle   string `json:"user_title" validate:"max=255"`
	CompanyName string `json:"company_name" validate:"max=255"`
	Rating      int    `json:"rating" validate:"omitempty,min=1,max=5"`
}

type UpdateTestimonialRequest struct {
	Content     *string `json:"content" validate:"omitempty,min=20,max=1000"`
	UserTitle   *string `json:"user_title" validate:"omitempty,max=255"`
	CompanyName *string `json:"company_name" validate:"omitempty,max=255"`
	Rating      *int    `json:"rating" validate:"omitempty,min=1,max=5"`
}
