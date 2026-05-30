package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupApplicationRoutes(router fiber.Router, applicationHandler *handlers.ApplicationHandler) {
	router.Get("/applications", applicationHandler.GetMyApplications)
	router.Get("/applications/:id", applicationHandler.GetApplicationDetails)
	router.Post("/jobs/:jobId/apply", applicationHandler.SubmitApplication)
	router.Put("/applications/:id/withdraw", applicationHandler.WithdrawApplication)
}