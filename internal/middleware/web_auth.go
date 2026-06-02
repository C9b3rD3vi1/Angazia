package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

func WebAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("access_token")
		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}
		if token == "" {
			return c.Redirect("/login", fiber.StatusFound)
		}

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			return c.Redirect("/login", fiber.StatusFound)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)
		c.Locals("user_email", claims.Email)

		return c.Next()
	}
}

func WebRequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role")
		if userRole == nil {
			return c.Redirect("/login", fiber.StatusFound)
		}
		roleStr, ok := userRole.(string)
		if !ok {
			return c.Redirect("/login", fiber.StatusFound)
		}
		for _, role := range roles {
			if roleStr == role {
				return c.Next()
			}
		}
		return c.Redirect("/login", fiber.StatusFound)
	}
}

func WebGuestMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("access_token")
		if token != "" {
			if claims, err := utils.ValidateJWT(token); err == nil {
				switch claims.Role {
				case "employee":
					return c.Redirect("/employee/dashboard", fiber.StatusFound)
				case "employer":
					return c.Redirect("/employer/dashboard", fiber.StatusFound)
				case "admin":
					return c.Redirect("/admin/dashboard", fiber.StatusFound)
				}
			}
		}
		return c.Next()
	}
}
