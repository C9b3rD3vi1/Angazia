package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
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
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req struct {
		Method      string `json:"method" validate:"oneof=app sms email"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	setup, err := h.twoFAService.InitiateSetup(c.Context(), userID.(string), req.Method, req.PhoneNumber, req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "setup_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    setup,
	})
}

// VerifySetup verifies and enables 2FA
func (h *TwoFAHandler) VerifySetup(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req models.TwoFAVerify
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	if err := h.twoFAService.VerifySetup(c.Context(), userID.(string), req.Code, req.Method); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "verification_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "2FA enabled successfully",
	})
}

// Disable disables 2FA
func (h *TwoFAHandler) Disable(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	var req models.TwoFADisable
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}
	
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	
	if err := h.twoFAService.Disable(c.Context(), userID.(string), req.Code, req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "disable_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "2FA disabled successfully",
	})
}

// GetStatus returns 2FA status
func (h *TwoFAHandler) GetStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	enabled, err := h.twoFAService.IsEnabled(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "status_failed",
			Message: err.Error(),
		})
	}
	
	method, _ := h.twoFAService.GetMethod(c.Context(), userID.(string))
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"enabled": enabled,
			"method":  method,
		},
	})
}

// GenerateBackupCodes generates new backup codes
func (h *TwoFAHandler) GenerateBackupCodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	
	codes, err := h.twoFAService.GenerateBackupCodes(c.Context(), userID.(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "generation_failed",
			Message: err.Error(),
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"backup_codes": codes,
		},
	})
}

// GetBackupCodes returns metadata about backup codes (cannot retrieve actual codes)
func (h *TwoFAHandler) GetBackupCodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"message":      "Backup codes cannot be retrieved for security reasons.",
			"action":       "Use the POST /auth/2fa/backup-codes/generate endpoint to generate new backup codes.",
		},
	})
}

func (h *TwoFAHandler) InitiateRecovery(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}

	if err := h.twoFAService.Initiate2FARecovery(c.Context(), userID.(string)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "recovery_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Recovery email sent. Check your email for the recovery link.",
	})
}

func (h *TwoFAHandler) CompleteRecovery(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	token := c.Query("token")

	if userID == "" || token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "User ID and recovery token are required",
		})
	}

	if err := h.twoFAService.Complete2FARecovery(c.Context(), userID, token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "recovery_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "2FA has been disabled. Please set up 2FA again.",
	})
}