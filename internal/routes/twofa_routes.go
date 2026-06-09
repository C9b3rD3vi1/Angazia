package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func SetupTwoFARoutes(api fiber.Router, twoFAHandler *handlers.TwoFAHandler) {
	twoFA := api.Group("/auth/2fa", middleware.AuthMiddleware())

	twoFA.Post("/setup", twoFAHandler.Setup)
	twoFA.Post("/verify", twoFAHandler.VerifySetup)
	twoFA.Post("/disable", twoFAHandler.Disable)
	twoFA.Get("/status", twoFAHandler.GetStatus)
	twoFA.Post("/backup-codes/generate", twoFAHandler.GenerateBackupCodes)
	twoFA.Get("/backup-codes", twoFAHandler.GetBackupCodes)
	twoFA.Post("/recovery", twoFAHandler.InitiateRecovery)
	twoFA.Get("/recover", twoFAHandler.CompleteRecovery)
	twoFA.Post("/login-verify", twoFAHandler.LoginVerify)
}

func SetupTwoFAGlobalMiddleware(app *fiber.App, twoFAService services.TwoFAService) {
	app.Use(middleware.TwoFAMiddleware(twoFAService))
}
