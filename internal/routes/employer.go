package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupEmployerRoutes(
	router fiber.Router,
	employerHandler *handlers.EmployerHandler,
	jobHandler *handlers.JobHandler,
) {
	// Dashboard
	router.Get("/dashboard", employerHandler.GetDashboard)
	router.Get("/analytics", employerHandler.GetAnalytics)
	router.Get("/stats", employerHandler.GetStats)
	
	// Company profile
	router.Get("/company", employerHandler.GetCompanyProfile)
	router.Put("/company", employerHandler.UpdateCompanyProfile)
	router.Post("/company/logo", employerHandler.UploadLogo)
	
	// Job management
	router.Get("/jobs", jobHandler.GetMyJobs)
	router.Post("/jobs", jobHandler.PostJob)
	router.Get("/jobs/:id", jobHandler.GetJobForEmployer)
	router.Put("/jobs/:id", jobHandler.UpdateJob)
	router.Delete("/jobs/:id", jobHandler.DeleteJob)
	router.Post("/jobs/:id/duplicate", jobHandler.DuplicateJob)
	router.Post("/jobs/:id/close", jobHandler.CloseJob)
	router.Post("/jobs/:id/feature", jobHandler.FeatureJob)
	
	// Application management
	router.Get("/jobs/:jobId/applications", employerHandler.GetJobApplications)
	router.Get("/applications/:id", employerHandler.GetApplicationDetails)
	router.Put("/applications/:id/status", employerHandler.UpdateApplicationStatus)
	router.Post("/applications/:id/note", employerHandler.AddApplicationNote)
	router.Post("/applications/:id/shortlist", employerHandler.ShortlistApplication)
	router.Post("/applications/:id/reject", employerHandler.RejectApplication)
	router.Post("/applications/:id/hire", employerHandler.HireCandidate)
	
	// Talent pool
	router.Get("/talent-pools", employerHandler.GetTalentPools)
	router.Post("/talent-pools", employerHandler.CreateTalentPool)
	router.Put("/talent-pools/:id", employerHandler.UpdateTalentPool)
	router.Delete("/talent-pools/:id", employerHandler.DeleteTalentPool)
	router.Post("/talent-pools/:id/candidates", employerHandler.AddCandidateToPool)
	router.Delete("/talent-pools/:id/candidates/:employeeId", employerHandler.RemoveCandidateFromPool)
	
	// Billing
	router.Get("/billing", employerHandler.GetBillingInfo)
	router.Post("/billing/subscribe", employerHandler.SubscribeToPlan)
}