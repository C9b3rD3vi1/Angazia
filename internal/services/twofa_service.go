package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

const (
	max2FAAttempts          = 5
	rateLimitWindow         = 15 * time.Minute
	lockoutDuration         = 15 * time.Minute
	codeExpiry              = 5 * time.Minute
	trustedDeviceExpiry     = 30 * 24 * time.Hour
	backupCodeCount         = 10
)

type backupCodeEntry struct {
	Hash string `json:"hash"`
	Used bool   `json:"used"`
}

type TwoFAService interface {
	InitiateSetup(ctx context.Context, userID, method, phoneNumber, email string) (*models.TwoFASetup, error)
	VerifySetup(ctx context.Context, userID string, code string, method string) error
	Disable(ctx context.Context, userID, code, password string) error

	VerifyCode(ctx context.Context, userID, code string, deviceID string) (bool, error)
	VerifyBackupCode(ctx context.Context, userID, code string) (bool, error)

	GenerateBackupCodes(ctx context.Context, userID string) ([]string, error)
	GetBackupCodes(ctx context.Context, userID string) ([]string, error)

	IsEnabled(ctx context.Context, userID string) (bool, error)
	GetMethod(ctx context.Context, userID string) (string, error)

	SendSMSCode(ctx context.Context, phoneNumber string) error
	SendEmailCode(ctx context.Context, email string) error

	IsDeviceTrusted(ctx context.Context, userID, deviceID string) bool

	Initiate2FARecovery(ctx context.Context, userID string) error
	Complete2FARecovery(ctx context.Context, userID, token string) error
}

type TwoFAServiceImpl struct {
	cfg          *config.Config
	twoFARepo    repository.TwoFARepository
	userRepo     repository.UserRepository
	smsProvider  SMSProvider
	emailService EmailService
	redisClient  *redis.Client
}

func NewTwoFAService(
	cfg *config.Config,
	twoFARepo repository.TwoFARepository,
	userRepo repository.UserRepository,
	smsProvider SMSProvider,
	emailService EmailService,
	redisClient *redis.Client,
) TwoFAService {
	return &TwoFAServiceImpl{
		cfg:          cfg,
		twoFARepo:    twoFARepo,
		userRepo:     userRepo,
		smsProvider:  smsProvider,
		emailService: emailService,
		redisClient:  redisClient,
	}
}

func (s *TwoFAServiceImpl) InitiateSetup(ctx context.Context, userID, method, phoneNumber, email string) (*models.TwoFASetup, error) {
	existing, _ := s.twoFARepo.GetByUserID(ctx, userID)
	if existing != nil && existing.IsEnabled {
		return nil, fmt.Errorf("2FA already enabled")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.AppName,
		AccountName: userID,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	backupCodes := s.generateBackupCodes()
	hashedCodes := s.hashBackupCodes(backupCodes)

	encryptedSecret, err := s.encrypt(key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	backupCodesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	var backupCodesArray models.JSONArray
	json.Unmarshal(backupCodesJSON, &backupCodesArray)

	secret := &models.TwoFASecret{
		UserID:      userID,
		Secret:      encryptedSecret,
		Method:      method,
		PhoneNumber: phoneNumber,
		Email:       email,
		BackupCodes: backupCodesArray,
		IsEnabled:   false,
		IsVerified:  false,
	}

	if err := s.twoFARepo.Upsert(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to save secret: %w", err)
	}

	s.logEvent(ctx, userID, "setup_initiated", method)

	qrCode, err := s.generateQRCode(key.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return &models.TwoFASetup{
		Secret:      key.Secret(),
		QRCodeURL:   qrCode,
		BackupCodes: backupCodes,
		Method:      method,
	}, nil
}

func (s *TwoFAServiceImpl) VerifySetup(ctx context.Context, userID string, code string, method string) error {
	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA setup not found: %w", err)
	}

	decryptedSecret, err := s.decrypt(secret.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	if !totp.Validate(code, decryptedSecret) {
		return fmt.Errorf("invalid verification code")
	}

	secret.IsEnabled = true
	secret.IsVerified = true
	secret.Method = method

	if err := s.twoFARepo.Update(ctx, secret); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	s.logEvent(ctx, userID, "enabled", method)

	if method == "email" && secret.Email != "" {
		s.emailService.SendTwoFAEnabled(secret.Email)
	}

	return nil
}

func (s *TwoFAServiceImpl) Disable(ctx context.Context, userID, code, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}

	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not enabled")
	}

	if len(code) == 8 {
		if !s.verifyBackupCode(secret, code) {
			return fmt.Errorf("invalid backup code")
		}
	} else {
		decryptedSecret, err := s.decrypt(secret.Secret)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret: %w", err)
		}
		if !totp.Validate(code, decryptedSecret) {
			return fmt.Errorf("invalid verification code")
		}
	}

	secret.IsEnabled = false
	secret.IsVerified = false

	if err := s.twoFARepo.Update(ctx, secret); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	s.logEvent(ctx, userID, "disabled", secret.Method)
	s.clearTrustedDevices(ctx, userID)

	if secret.Email != "" {
		s.emailService.SendTwoFADisabled(secret.Email)
	}

	return nil
}

func (s *TwoFAServiceImpl) VerifyCode(ctx context.Context, userID, code string, deviceID string) (bool, error) {
	if err := s.checkRateLimit(ctx, userID); err != nil {
		return false, err
	}

	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("2FA not enabled")
	}

	if !secret.IsEnabled {
		return false, fmt.Errorf("2FA not enabled")
	}

	var valid bool

	if len(code) == 8 {
		valid = s.verifyBackupCode(secret, code)
	} else if len(code) == 6 {
		decryptedSecret, err := s.decrypt(secret.Secret)
		if err == nil {
			valid = totp.Validate(code, decryptedSecret)
		}
	} else {
		valid = s.verifyStoredCode(ctx, userID, code)
	}

	if valid {
		s.clearRateLimit(ctx, userID)
		now := time.Now()
		secret.LastUsedAt = &now
		s.twoFARepo.Update(ctx, secret)

		if deviceID != "" {
			s.recordTrustedDevice(ctx, userID, deviceID)
		}
		return true, nil
	}

	s.recordFailedAttempt(ctx, userID)
	s.logEvent(ctx, userID, "failed_attempt", "")
	return false, nil
}

func (s *TwoFAServiceImpl) VerifyBackupCode(ctx context.Context, userID, code string) (bool, error) {
	if err := s.checkRateLimit(ctx, userID); err != nil {
		return false, err
	}

	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	valid := s.verifyBackupCode(secret, code)
	if valid {
		s.clearRateLimit(ctx, userID)
	} else {
		s.recordFailedAttempt(ctx, userID)
	}
	return valid, nil
}

func (s *TwoFAServiceImpl) GenerateBackupCodes(ctx context.Context, userID string) ([]string, error) {
	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	rawCodes := s.generateBackupCodes()
	hashedCodes := s.hashBackupCodes(rawCodes)

	codesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	var codesArray models.JSONArray
	json.Unmarshal(codesJSON, &codesArray)
	secret.BackupCodes = codesArray
	secret.RecoveryCodesUsed = 0

	if err := s.twoFARepo.Update(ctx, secret); err != nil {
		return nil, err
	}

	s.logEvent(ctx, userID, "backup_codes_generated", "")
	return rawCodes, nil
}

func (s *TwoFAServiceImpl) GetBackupCodes(ctx context.Context, userID string) ([]string, error) {
	return nil, fmt.Errorf("backup codes cannot be retrieved for security reasons; use the generate endpoint to create new codes")
}

func (s *TwoFAServiceImpl) IsEnabled(ctx context.Context, userID string) (bool, error) {
	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, nil
	}
	return secret.IsEnabled, nil
}

func (s *TwoFAServiceImpl) GetMethod(ctx context.Context, userID string) (string, error) {
	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	return secret.Method, nil
}

func (s *TwoFAServiceImpl) SendSMSCode(ctx context.Context, phoneNumber string) error {
	code := s.generateSMSCode()
	now := time.Now().Unix()
	codeKey := fmt.Sprintf("2fa:code:sms:%s", phoneNumber)
	payload := fmt.Sprintf("%s:%d", code, now)

	if s.redisClient != nil {
		if err := s.redisClient.Set(ctx, codeKey, payload, codeExpiry).Err(); err != nil {
			return fmt.Errorf("failed to store SMS code: %w", err)
		}
	}

	return s.smsProvider.Send(phoneNumber, fmt.Sprintf("Your %s verification code is: %s. Expires in 5 minutes.", s.cfg.AppName, code))
}

func (s *TwoFAServiceImpl) SendEmailCode(ctx context.Context, email string) error {
	code := s.generateEmailCode()
	now := time.Now().Unix()
	codeKey := fmt.Sprintf("2fa:code:email:%s", email)
	payload := fmt.Sprintf("%s:%d", code, now)

	if s.redisClient != nil {
		if err := s.redisClient.Set(ctx, codeKey, payload, codeExpiry).Err(); err != nil {
			return fmt.Errorf("failed to store email code: %w", err)
		}
	}

	return s.emailService.SendTwoFACode(email, code)
}

func (s *TwoFAServiceImpl) generateBackupCodes() []string {
	codes := make([]string, backupCodeCount)
	for i := 0; i < backupCodeCount; i++ {
		bytes := make([]byte, 4)
		rand.Read(bytes)
		codes[i] = fmt.Sprintf("%08X", bytes)
	}
	return codes
}

func (s *TwoFAServiceImpl) hashBackupCodes(codes []string) []backupCodeEntry {
	entries := make([]backupCodeEntry, len(codes))
	for i, code := range codes {
		hash := sha256.Sum256([]byte(code))
		entries[i] = backupCodeEntry{
			Hash: hex.EncodeToString(hash[:]),
			Used: false,
		}
	}
	return entries
}

func (s *TwoFAServiceImpl) verifyBackupCode(secret *models.TwoFASecret, code string) bool {
	entriesBytes, err := json.Marshal(secret.BackupCodes)
	if err != nil {
		return false
	}

	var entries []backupCodeEntry
	if err := json.Unmarshal(entriesBytes, &entries); err != nil {
		return false
	}

	inputHash := sha256.Sum256([]byte(code))
	inputHashHex := hex.EncodeToString(inputHash[:])

	for i, entry := range entries {
		if !entry.Used && hmac.Equal([]byte(entry.Hash), []byte(inputHashHex)) {
			entries[i].Used = true
			updatedJSON, err := json.Marshal(entries)
			if err != nil {
				return false
			}
			var updatedArray models.JSONArray
			json.Unmarshal(updatedJSON, &updatedArray)
			secret.BackupCodes = updatedArray
			secret.RecoveryCodesUsed++
			s.twoFARepo.Update(context.Background(), secret)
			return true
		}
	}

	return false
}

func (s *TwoFAServiceImpl) verifyStoredCode(ctx context.Context, userID, code string) bool {
	if s.redisClient == nil {
		return false
	}

	patterns := []string{
		fmt.Sprintf("2fa:code:sms:*"),
		fmt.Sprintf("2fa:code:email:*"),
	}

	for _, pattern := range patterns {
		iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			key := iter.Val()
			if !strings.Contains(key, userID) {
				stored, err := s.redisClient.Get(ctx, key).Result()
				if err != nil {
					continue
				}
				parts := strings.SplitN(stored, ":", 2)
				if len(parts) == 2 && parts[0] == code {
					s.redisClient.Del(ctx, key)
					return true
				}
			}
		}
	}

	return false
}

func (s *TwoFAServiceImpl) generateSMSCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return fmt.Sprintf("%06d", int(bytes[0])<<16|int(bytes[1])<<8|int(bytes[2]))
}

func (s *TwoFAServiceImpl) generateEmailCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%08d", int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))
}

func (s *TwoFAServiceImpl) generateQRCode(url string) (string, error) {
	key, err := otp.NewKeyFromURL(url)
	if err != nil {
		return "", err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *TwoFAServiceImpl) encryptionKey() []byte {
	key := s.cfg.TwoFAEncryptionKey
	if key == "" {
		key = s.cfg.JWTSecret
	}
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

func (s *TwoFAServiceImpl) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey())
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(nonce) + ":" + hex.EncodeToString(ciphertext), nil
}

func (s *TwoFAServiceImpl) decrypt(encrypted string) (string, error) {
	parts := strings.SplitN(encrypted, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted format")
	}

	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid nonce: %w", err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey())
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

func (s *TwoFAServiceImpl) Initiate2FARecovery(ctx context.Context, userID string) error {
	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not enabled")
	}

	if !secret.IsEnabled {
		return fmt.Errorf("2FA is not enabled")
	}

	if s.redisClient == nil {
		return fmt.Errorf("recovery unavailable: no Redis connection")
	}

	recoveryToken := uuid.New().String()
	tokenKey := fmt.Sprintf("2fa:recovery:%s:%s", userID, recoveryToken)

	if err := s.redisClient.Set(ctx, tokenKey, "1", 15*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to create recovery token: %w", err)
	}

	recoveryLink := fmt.Sprintf("%s/auth/2fa/recover?user_id=%s&token=%s", s.cfg.AppURL, userID, recoveryToken)

	if secret.Email != "" {
		s.emailService.SendTwoFARecoveryEmail(secret.Email, recoveryLink)
	}

	s.logEvent(ctx, userID, "recovery_initiated", secret.Method)
	return nil
}

func (s *TwoFAServiceImpl) Complete2FARecovery(ctx context.Context, userID, token string) error {
	if s.redisClient == nil {
		return fmt.Errorf("recovery unavailable: no Redis connection")
	}

	tokenKey := fmt.Sprintf("2fa:recovery:%s:%s", userID, token)
	exists, err := s.redisClient.Exists(ctx, tokenKey).Result()
	if err != nil || exists == 0 {
		return fmt.Errorf("invalid or expired recovery token")
	}

	secret, err := s.twoFARepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("2FA not enabled")
	}

	secret.IsEnabled = false
	secret.IsVerified = false

	if err := s.twoFARepo.Update(ctx, secret); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	s.redisClient.Del(ctx, tokenKey)
	s.clearTrustedDevices(ctx, userID)
	s.logEvent(ctx, userID, "recovery_completed", secret.Method)

	if secret.Email != "" {
		s.emailService.SendTwoFADisabled(secret.Email)
	}

	return nil
}

func (s *TwoFAServiceImpl) logEvent(ctx context.Context, userID, action, method string) {
	entry := &models.TwoFAAuditLog{
		UserID: userID,
		Action: action,
		Method: method,
		Metadata: models.JSONMap{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	s.twoFARepo.LogEvent(ctx, entry)
}

func (s *TwoFAServiceImpl) recordTrustedDevice(ctx context.Context, userID, deviceID string) {
	if s.redisClient == nil {
		return
	}

	key := fmt.Sprintf("2fa:trusted:%s:%s", userID, deviceID)
	s.redisClient.Set(ctx, key, "1", trustedDeviceExpiry)
}

func (s *TwoFAServiceImpl) IsDeviceTrusted(ctx context.Context, userID, deviceID string) bool {
	if s.redisClient == nil || deviceID == "" {
		return false
	}

	key := fmt.Sprintf("2fa:trusted:%s:%s", userID, deviceID)
	exists, err := s.redisClient.Exists(ctx, key).Result()
	return err == nil && exists == 1
}

func (s *TwoFAServiceImpl) clearTrustedDevices(ctx context.Context, userID string) {
	if s.redisClient == nil {
		return
	}

	pattern := fmt.Sprintf("2fa:trusted:%s:*", userID)
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
}

func (s *TwoFAServiceImpl) checkRateLimit(ctx context.Context, userID string) error {
	if s.redisClient == nil {
		return nil
	}

	lockoutKey := fmt.Sprintf("2fa:lockout:%s", userID)
	locked, err := s.redisClient.Exists(ctx, lockoutKey).Result()
	if err == nil && locked == 1 {
		return fmt.Errorf("account temporarily locked due to too many failed 2FA attempts; try again in 15 minutes")
	}

	return nil
}

func (s *TwoFAServiceImpl) recordFailedAttempt(ctx context.Context, userID string) {
	if s.redisClient == nil {
		return
	}

	failKey := fmt.Sprintf("2fa:fail:%s", userID)
	count, err := s.redisClient.Incr(ctx, failKey).Result()
	if err != nil {
		return
	}

	if count == 1 {
		s.redisClient.Expire(ctx, failKey, rateLimitWindow)
	}

	if count >= max2FAAttempts {
		lockoutKey := fmt.Sprintf("2fa:lockout:%s", userID)
		s.redisClient.Set(ctx, lockoutKey, "1", lockoutDuration)
		s.redisClient.Del(ctx, failKey)
	}
}

func (s *TwoFAServiceImpl) clearRateLimit(ctx context.Context, userID string) {
	if s.redisClient == nil {
		return
	}

	failKey := fmt.Sprintf("2fa:fail:%s", userID)
	lockoutKey := fmt.Sprintf("2fa:lockout:%s", userID)
	s.redisClient.Del(ctx, failKey, lockoutKey)
}
