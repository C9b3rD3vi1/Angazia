package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupCompanyRoutes(router fiber.Router, companyHandler *handlers.CompanyHandler) {
	// Public company routes (no authentication)
	router.Get("/companies/:id", companyHandler.GetPublicCompanyProfile)
	router.Get("/companies/:id/badges", companyHandler.GetCompanyBadges)
	router.Get("/companies/:id/reviews", companyHandler.GetCompanyReviews)
	router.Get("/companies/:id/reviews/stats", companyHandler.GetReviewStats)
	
	// Protected routes (authentication required)
	protected := router.Group("/", middleware.AuthMiddleware())
	
	// Employer company management routes
	employer := protected.Group("/employer", middleware.RequireRole("employer"))
	employer.Get("/company", companyHandler.GetCompanyProfile)
	employer.Put("/company", companyHandler.UpdateCompanyProfile)
	employer.Post("/company/logo", companyHandler.UploadCompanyLogo)
	employer.Post("/company/verify", companyHandler.SubmitVerification)
	employer.Get("/company/verification", companyHandler.GetVerificationStatus)
	employer.Get("/company/badges", companyHandler.GetCompanyBadges)
	
	// Team management
	employer.Get("/team", companyHandler.GetTeamMembers)
	employer.Post("/team/invite", companyHandler.InviteTeamMember)
	employer.Put("/team/:memberId/role", companyHandler.UpdateTeamMemberRole)
	employer.Delete("/team/:memberId", companyHandler.RemoveTeamMember)
	employer.Get("/team/invitations", companyHandler.ListPendingInvitations)
	employer.Delete("/team/invitations/:invitationId", companyHandler.CancelInvitation)
	
	// Analytics
	employer.Get("/analytics", companyHandler.GetCompanyAnalytics)
	
	// Review submission (authenticated users)
	protected.Post("/companies/:id/reviews", companyHandler.SubmitReview)
	
	// Invitation acceptance (no role required, just authentication)
	protected.Post("/invitations/:token/accept", companyHandler.AcceptInvitation)
}