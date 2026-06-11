package models

import (
	"time"
)

// TwoFASecret stores user's 2FA configuration
type TwoFASecret struct {
	ID              string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string     `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	Secret          string     `json:"secret" gorm:"not null"` // Encrypted TOTP secret
	Method          string     `json:"method" gorm:"size:20;default:'app'"` // app, sms, email
	PhoneNumber     string     `json:"phone_number,omitempty" gorm:"size:20"`
	Email           string     `json:"email,omitempty" gorm:"size:255"`
	BackupCodes     JSONArray  `json:"backup_codes" gorm:"type:jsonb"`
	IsEnabled       bool       `json:"is_enabled" gorm:"default:false"`
	IsVerified      bool       `json:"is_verified" gorm:"default:false"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	RecoveryCodesUsed int      `json:"recovery_codes_used" gorm:"default:0"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	User            *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TwoFASetup represents the setup response for 2FA
type TwoFASetup struct {
	Secret        string   `json:"secret"`
	QRCodeURL     string   `json:"qr_code"`
	BackupCodes   []string `json:"backup_codes"`
	Method        string   `json:"method"`
}

// TwoFAVerify represents verification request
type TwoFAVerify struct {
	Code     string `json:"code" validate:"required"`
	Method   string `json:"method"`
	DeviceID string `json:"device_id,omitempty"`
}

// TwoFADisable represents disable request
type TwoFADisable struct {
	Code     string `json:"code" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TwoFABackupCode represents a single backup code
type TwoFABackupCode struct {
	Code      string    `json:"code"`
	Used      bool      `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type TwoFAAuditLog struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Action    string    `json:"action" gorm:"size:50;not null"` // enabled, disabled, verified, backup_codes_generated, failed_attempt
	Method    string    `json:"method" gorm:"size:20"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Metadata  JSONMap   `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

func (TwoFASecret) TableName() string {
	return "two_fa_secrets"
}

func (TwoFAAuditLog) TableName() string {
	return "two_fa_audit_logs"
}