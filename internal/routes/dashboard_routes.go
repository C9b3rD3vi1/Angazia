package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupDashboardRoutes(router fiber.Router, dashboardHandler *handlers.DashboardHandler) {
	dash := router.Group("/employer/dashboard", middleware.AuthMiddleware(), middleware.RequireRole("employer"))
	dash.Get("", dashboardHandler.GetEmployerDashboard)
}
