package models

import (
	"time"
)

// Payment represents a payment transaction
type Payment struct {
	ID              string                 `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string                 `json:"user_id" gorm:"type:uuid;not null;index"`
	SubscriptionID  *string                `json:"subscription_id,omitempty" gorm:"type:uuid;index"`
	Amount          float64                `json:"amount" gorm:"not null"`
	Currency        string                 `json:"currency" gorm:"size:3;default:'KES'"`
	Status          string                 `json:"status" gorm:"size:50;default:'pending';index"`
	PaymentMethod   string                 `json:"payment_method" gorm:"size:50"`
	TransactionID   string                 `json:"transaction_id" gorm:"size:255;index"`
	Reference       string                 `json:"reference" gorm:"size:255;uniqueIndex"`
	Description     string                 `json:"description" gorm:"type:text"`
	Metadata        JSONMap                `json:"metadata" gorm:"type:jsonb"`
	ReceiptURL      string                 `json:"receipt_url" gorm:"size:512"`
	PaidAt          *time.Time             `json:"paid_at,omitempty"`
	RefundedAt      *time.Time             `json:"refunded_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// PaymentIntent represents a payment intent created with IntaSend
type PaymentIntent struct {
	ID              string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency" gorm:"default:'KES'"`
	PlanID          string    `json:"plan_id" gorm:"size:100"`
	Status          string    `json:"status" gorm:"size:50;default:'pending'"`
	InvoiceID       string    `json:"invoice_id" gorm:"size:255"`
	RedirectURL     string    `json:"redirect_url" gorm:"size:512"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// IntaSendWebhookPayload represents the webhook payload from IntaSend
type IntaSendWebhookPayload struct {
	Event       string                 `json:"event"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Signature   string                 `json:"signature"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceNumber   string         `json:"invoice_number" gorm:"size:50;uniqueIndex"`
	UserID          string         `json:"user_id" gorm:"type:uuid;not null;index"`
	SubscriptionID  *string        `json:"subscription_id,omitempty" gorm:"type:uuid;index"`
	PaymentID       *string        `json:"payment_id,omitempty" gorm:"type:uuid;index"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency" gorm:"default:'KES'"`
	Tax             float64        `json:"tax" gorm:"default:0"`
	Total           float64        `json:"total"`
	Status          string         `json:"status" gorm:"size:50;default:'pending'"`
	DueDate         time.Time      `json:"due_date"`
	PaidAt          *time.Time     `json:"paid_at,omitempty"`
	PDFURL          string         `json:"pdf_url" gorm:"size:512"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
	Payment      *Payment      `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// InvoiceItem represents an item on an invoice
type InvoiceItem struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceID   string    `json:"invoice_id" gorm:"type:uuid;not null;index"`
	Description string    `json:"description" gorm:"type:text;not null"`
	Quantity    int       `json:"quantity" gorm:"default:1"`
	UnitPrice   float64   `json:"unit_price"`
	Total       float64   `json:"total"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Payment) TableName() string {
	return "payments"
}

func (PaymentIntent) TableName() string {
	return "payment_intents"
}

func (Invoice) TableName() string {
	return "invoices"
}

func (InvoiceItem) TableName() string {
	return "invoice_items"
}