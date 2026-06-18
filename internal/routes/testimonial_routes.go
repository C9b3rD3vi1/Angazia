package routes

import (
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupTestimonialRoutes(router fiber.Router, h *handlers.TestimonialHandler) {
	// Public - approved testimonials
	router.Get("/testimonials", h.ListApproved)

	// Authenticated users - manage own testimonials
	protected := router.Group("/testimonials", middleware.AuthMiddleware())
	protected.Get("/mine", h.ListMy)
	protected.Post("/", h.Create)
	protected.Put("/:id", h.Update)
	protected.Delete("/:id", h.DeleteMy)
}

func SetupAdminTestimonialRoutes(router fiber.Router, h *handlers.TestimonialHandler) {
	admin := router.Group("/admin/testimonials", middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	admin.Get("/", h.AdminList)
	admin.Post("/:id/approve", h.AdminApprove)
	admin.Post("/:id/reject", h.AdminReject)
	admin.Post("/:id/feature", h.AdminToggleFeatured)
	admin.Delete("/:id", h.AdminDelete)
}
