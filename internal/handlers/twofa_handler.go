package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

type TwoFAHandler struct {
	twoFAService services.TwoFAService
	validator    *validator.Validate
}

func NewTwoFAHandler(twoFAService services.TwoFAService) *TwoFAHandler {
	return &TwoFAHandler{
		twoFAService: twoFAService,
		validator:    validator.New(),
	}
}

// Setup initiates 2FA setup
func (h *TwoFAHandler) Setup(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		Method      string `json:"method"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		req.Method = "app"
	}

	setup, err := h.twoFAService.InitiateSetup(c.Context(), userID.(string), req.Method, req.PhoneNumber, req.Email)
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}

	return utils.Success(c, setup)
}

// VerifySetup verifies and enables 2FA
func (h *TwoFAHandler) VerifySetup(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req models.TwoFAVerify
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.twoFAService.VerifySetup(c.Context(), userID.(string), req.Code, req.Method); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	backupCodes, _ := h.twoFAService.GenerateBackupCodes(c.Context(), userID.(string))
	
	return utils.SuccessWithMessage(c, "2FA enabled successfully", fiber.Map{
		"backup_codes": backupCodes,
	})
}

// Disable disables 2FA
func (h *TwoFAHandler) Disable(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	var req models.TwoFADisable
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	if err := h.twoFAService.Disable(c.Context(), userID.(string), req.Code, req.Password); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	
	return utils.SuccessWithMessage(c, "2FA disabled successfully", nil)
}

// GetStatus returns 2FA status
func (h *TwoFAHandler) GetStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	enabled, err := h.twoFAService.IsEnabled(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	method, _ := h.twoFAService.GetMethod(c.Context(), userID.(string))
	
	return utils.Success(c, fiber.Map{
		"enabled": enabled,
		"method":  method,
	})
}

// GenerateBackupCodes generates new backup codes
func (h *TwoFAHandler) GenerateBackupCodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}
	
	codes, err := h.twoFAService.GenerateBackupCodes(c.Context(), userID.(string))
	if err != nil {
		return utils.InternalServerError(c, err.Error())
	}
	
	return utils.Success(c, fiber.Map{
		"backup_codes": codes,
	})
}

// GetBackupCodes returns metadata about backup codes (cannot retrieve actual codes)
func (h *TwoFAHandler) GetBackupCodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	return utils.Success(c, fiber.Map{
		"message":      "Backup codes cannot be retrieved for security reasons.",
		"action":       "Use the POST /auth/2fa/backup-codes/generate endpoint to generate new backup codes.",
		"can_generate": true,
	})
}

func (h *TwoFAHandler) LoginVerify(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	var req struct {
		Code string `json:"code" validate:"required,min=6,max=8"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	deviceID := generateDeviceID()
	valid, err := h.twoFAService.VerifyCode(c.Context(), userID.(string), req.Code, deviceID)
	if err != nil || !valid {
		return utils.Unauthorized(c, "Invalid 2FA verification code")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "twofa_device_id",
		Value:    deviceID,
		HTTPOnly: true,
		Secure:   true,
		MaxAge:   30 * 24 * 3600,
		SameSite: "Strict",
		Path:     "/",
	})

	return utils.SuccessWithMessage(c, "2FA verification successful", fiber.Map{
		"device_id": deviceID,
	})
}

func generateDeviceID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "device_" + hex.EncodeToString(bytes)
}

func (h *TwoFAHandler) InitiateRecovery(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Unauthorized(c, "User not authenticated")
	}

	if err := h.twoFAService.Initiate2FARecovery(c.Context(), userID.(string)); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "Recovery email sent. Check your email for the recovery link.", nil)
}

func (h *TwoFAHandler) CompleteRecovery(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	token := c.Query("token")

	if userID == "" || token == "" {
		return utils.BadRequest(c, "User ID and recovery token are required")
	}

	if err := h.twoFAService.Complete2FARecovery(c.Context(), userID, token); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.SuccessWithMessage(c, "2FA has been disabled. Please set up 2FA again.", nil)
}
