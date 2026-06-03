package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/repository"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

type UnsubscribeHandler struct {
	unsubscribeRepo repository.UnsubscribeRepository
	emailService    services.EmailService
}

func NewUnsubscribeHandler(unsubscribeRepo repository.UnsubscribeRepository, emailService services.EmailService) *UnsubscribeHandler {
	return &UnsubscribeHandler{
		unsubscribeRepo: unsubscribeRepo,
		emailService:    emailService,
	}
}

// Unsubscribe handles email unsubscribe requests
func (h *UnsubscribeHandler) Unsubscribe(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return utils.BadRequest(c, "Unsubscribe token is required")
	}

	// Validate token
	unsubscribeToken, err := h.unsubscribeRepo.GetToken(c.Context(), token)
	if err != nil {
		return utils.NotFound(c, "Token")
	}

	// Deactivate token
	if err := h.unsubscribeRepo.DeactivateToken(c.Context(), token); err != nil {
		return utils.InternalServerError(c, "Failed to unsubscribe. Please try again later.")
	}

	// Also unsubscribe from all email types for this user
	if err := h.unsubscribeRepo.UnsubscribeAll(c.Context(), unsubscribeToken.Email); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to unsubscribe all for %s: %v\n", unsubscribeToken.Email, err)
	}

	// Return success page
	return utils.SuccessWithMessage(c, "Successfully unsubscribed from email notifications", nil)
}

// UpdatePreferences updates user email preferences
func (h *UnsubscribeHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return utils.Unauthorized(c, "")
	}

	var req struct {
		JobAlerts          *bool   `json:"job_alerts"`
		ApplicationUpdates *bool   `json:"application_updates"`
		MarketingEmails    *bool   `json:"marketing_emails"`
		SecurityAlerts     *bool   `json:"security_alerts"`
		Newsletter         *bool   `json:"newsletter"`
		DigestFrequency    string  `json:"digest_frequency"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request")
	}

	updates := make(map[string]interface{})
	if req.JobAlerts != nil {
		updates["job_alerts"] = *req.JobAlerts
	}
	if req.ApplicationUpdates != nil {
		updates["application_updates"] = *req.ApplicationUpdates
	}
	if req.MarketingEmails != nil {
		updates["marketing_emails"] = *req.MarketingEmails
	}
	if req.SecurityAlerts != nil {
		updates["security_alerts"] = *req.SecurityAlerts
	}
	if req.Newsletter != nil {
		updates["newsletter"] = *req.Newsletter
	}
	if req.DigestFrequency != "" {
		updates["digest_frequency"] = req.DigestFrequency
	}

	if err := h.unsubscribeRepo.UpdatePreferences(c.Context(), userID, updates); err != nil {
		return utils.InternalServerError(c, "Failed to update preferences")
	}

	return utils.SuccessWithMessage(c, "Email preferences updated successfully", nil)
}

// GetPreferences returns user email preferences
func (h *UnsubscribeHandler) GetPreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return utils.Unauthorized(c, "")
	}

	prefs, err := h.unsubscribeRepo.GetPreferences(c.Context(), userID)
	if err != nil {
		return utils.NotFound(c, "Preferences")
	}

	return utils.Success(c, fiber.Map{
		"preferences": prefs,
	})
}
