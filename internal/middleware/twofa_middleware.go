package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func TwoFAMiddleware(twoFAService services.TwoFAService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Next()
		}

		enabled, err := twoFAService.IsEnabled(c.Context(), userID.(string))
		if err != nil || !enabled {
			return c.Next()
		}

		twoFADeviceID := c.Cookies("twofa_device_id")
		if twoFADeviceID != "" && twoFAService.IsDeviceTrusted(c.Context(), userID.(string), twoFADeviceID) {
			return c.Next()
		}

		twoFACode := c.Get("X-2FA-Code")
		if twoFACode == "" {
			return utils.Unauthorized(c, "2FA verification required")
		}

		deviceID := generateDeviceID()
		valid, err := twoFAService.VerifyCode(c.Context(), userID.(string), twoFACode, deviceID)
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
		})

		return c.Next()
	}
}

func generateDeviceID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "device_" + hex.EncodeToString(bytes)
}
