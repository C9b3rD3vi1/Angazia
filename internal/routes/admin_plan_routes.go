// internal/routes/admin_plan_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupAdminPlanRoutes(router fiber.Router, adminPlanHandler *handlers.AdminPlanHandler) {
	// All plan management routes require admin authentication
	admin := router.Group("/admin/plans", middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	
	// Plan CRUD
	admin.Post("/", adminPlanHandler.CreatePlan)
	admin.Get("/", adminPlanHandler.GetPlans)
	admin.Get("/:id", adminPlanHandler.GetPlan)
	admin.Put("/:id", adminPlanHandler.UpdatePlan)
	admin.Delete("/:id", adminPlanHandler.DeletePlan)
	admin.Post("/:id/toggle", adminPlanHandler.TogglePlanActive)
}