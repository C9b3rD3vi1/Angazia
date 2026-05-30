package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
	"github.com/Angazia/internal/middleware"
)

// SetupJobRoutes configures job-related endpoints
func SetupJobRoutes(router fiber.Router) {
	router.Get("/", handlers.ListJobs)
	router.Get("/:id", handlers.GetJobDetails)
	router.Post("/:id/save", middleware.RequireRole("employee"), handlers.SaveJob)
	router.Delete("/:id/save", middleware.RequireRole("employee"), handlers.UnsaveJob)
	router.Get("/saved/list", middleware.RequireRole("employee"), handlers.GetSavedJobs)
}