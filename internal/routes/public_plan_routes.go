// internal/routes/public_plan_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
)

func SetupPublicPlanRoutes(router fiber.Router, adminPlanHandler *handlers.AdminPlanHandler) {
	// Public endpoints (no authentication required)
	router.Get("/plans", adminPlanHandler.GetPlans)
	router.Get("/plans/:id", adminPlanHandler.GetPlan)
}