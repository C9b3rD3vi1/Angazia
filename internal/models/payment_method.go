package models

import (
	"time"
)

type PaymentMethod struct {
	ID             string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Type           string    `json:"type" gorm:"size:50;not null"` // mpesa, card, bank
	Provider       string    `json:"provider" gorm:"size:50"`       // intasend, stripe
	Last4          string    `json:"last4" gorm:"size:4"`
	PhoneNumber    string    `json:"phone_number" gorm:"size:20"`
	CardBrand      string    `json:"card_brand" gorm:"size:50"`
	ExpiryMonth    int       `json:"expiry_month"`
	ExpiryYear     int       `json:"expiry_year"`
	IsDefault      bool      `json:"is_default" gorm:"default:false"`
	IsValid        bool      `json:"is_valid" gorm:"default:true"`
	Token          string    `json:"-" gorm:"size:512"`
	Metadata       JSONMap   `json:"metadata" gorm:"type:jsonb"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	User           *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (PaymentMethod) TableName() string {
	return "payment_methods"
}
