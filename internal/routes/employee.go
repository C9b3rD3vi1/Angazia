package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Angazia/internal/handlers"
)

func setupEmployeeRoutes(
	router fiber.Router,
	employeeHandler *handlers.EmployeeHandler,
	jobHandler *handlers.JobHandler,
	applicationHandler *handlers.ApplicationHandler,
) {
	// Dashboard
	router.Get("/dashboard", employeeHandler.GetDashboard)
	router.Get("/stats", employeeHandler.GetStats)
	
	// Profile management
	router.Get("/profile", employeeHandler.GetProfile)
	router.Put("/profile", employeeHandler.UpdateProfile)
	router.Post("/profile/upload-resume", employeeHandler.UploadResume)
	router.Delete("/profile/resume", employeeHandler.DeleteResume)
	
	// Skills management
	router.Post("/skills", employeeHandler.AddSkill)
	router.Delete("/skills/:skill", employeeHandler.RemoveSkill)
	router.Post("/skills/analyze", employeeHandler.AnalyzeSkills)
	
	// GitHub integration
	router.Post("/github/connect", employeeHandler.ConnectGitHub)
	router.Post("/github/disconnect", employeeHandler.DisconnectGitHub)
	router.Post("/github/sync", employeeHandler.SyncGitHub)
	router.Get("/github/profile", employeeHandler.GetGitHubProfile)
	router.Get("/github/repos", employeeHandler.GetGitHubRepos)
	router.Get("/github/contributions", employeeHandler.GetGitHubContributions)
	
	// Job matches
	router.Get("/matches", employeeHandler.GetMatches)
	router.Get("/matches/:id", employeeHandler.GetMatchDetails)
	router.Post("/matches/:id/feedback", employeeHandler.SubmitMatchFeedback)
	
	// Applications
	router.Get("/applications", applicationHandler.GetMyApplications)
	router.Get("/applications/:id", applicationHandler.GetApplicationDetails)
	router.Post("/applications/:id/withdraw", applicationHandler.WithdrawApplication)
	
	// Saved jobs
	router.Get("/saved-jobs", jobHandler.GetSavedJobs)
	
	// Settings
	router.Get("/settings", employeeHandler.GetSettings)
	router.Put("/settings", employeeHandler.UpdateSettings)
	router.Put("/settings/notifications", employeeHandler.UpdateNotificationSettings)
}