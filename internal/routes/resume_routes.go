package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupResumeRoutes(router fiber.Router, resumeHandler *handlers.ResumeHandler) {
	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())

	// Avatar upload (any authenticated user)
	protected.Post("/user/avatar", resumeHandler.UploadAvatar)
	
	// Employee profile routes
	employee := protected.Group("/employee", middleware.RequireRole("employee"))
	employee.Post("/resume/upload", resumeHandler.UploadResume)
	employee.Get("/profile/completion", resumeHandler.GetProfileCompletion)
	employee.Get("/skills/suggested", resumeHandler.GetSuggestedSkills)
	employee.Get("/profile/wizard", resumeHandler.GetProfileWizard)
}