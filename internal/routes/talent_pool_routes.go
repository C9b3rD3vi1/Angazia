package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
)

func SetupTalentPoolRoutes(router fiber.Router, talentPoolHandler *handlers.TalentPoolHandler) {
	// All routes require authentication and employer role
	protected := router.Group("/employer", middleware.AuthMiddleware(), middleware.RequireRole("employer"))
	
	// Talent pool management
	protected.Post("/talent-pools", talentPoolHandler.CreatePool)
	protected.Get("/talent-pools", talentPoolHandler.ListPools)
	protected.Get("/talent-pools/stats", talentPoolHandler.GetEmployerStats)
	protected.Get("/talent-pools/:id", talentPoolHandler.GetPool)
	protected.Put("/talent-pools/:id", talentPoolHandler.UpdatePool)
	protected.Delete("/talent-pools/:id", talentPoolHandler.DeletePool)
	protected.Get("/talent-pools/:id/stats", talentPoolHandler.GetPoolStats)
	protected.Get("/talent-pools/:id/search", talentPoolHandler.SearchCandidates)
	
	// Candidate management in talent pools
	protected.Get("/talent-pools/:id/candidates", talentPoolHandler.ListCandidates)
	protected.Post("/talent-pools/:id/candidates", talentPoolHandler.AddCandidate)
	protected.Put("/talent-pools/:poolId/candidates/:candidateId", talentPoolHandler.UpdateCandidate)
	protected.Delete("/talent-pools/:poolId/candidates/:candidateId", talentPoolHandler.RemoveCandidate)
	protected.Post("/talent-pools/:poolId/candidates/:candidateId/contact", talentPoolHandler.MarkContacted)
	protected.Post("/talent-pools/:poolId/candidates/:candidateId/hire", talentPoolHandler.MarkHired)
}