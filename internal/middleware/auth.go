package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.Unauthorized(c, "Authorization token is required")
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return utils.Unauthorized(c, "Token must be in Bearer format")
		}

		// Validate token
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			return utils.Unauthorized(c, err.Error())
		}

		// Set user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)
		c.Locals("user_email", claims.Email)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role")
		if userRole == nil {
			return utils.Unauthorized(c, "User role not found")
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				return c.Next()
			}
		}

		return utils.Forbidden(c, "Insufficient permissions")
	}
}
