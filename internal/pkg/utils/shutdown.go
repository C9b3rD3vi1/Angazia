package utils

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func WaitForShutdown(app *fiber.App) {
	// Create channel for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-shutdown

	log.Println("🛑 Shutting down server...")

	// Gracefully shutdown the server
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("✅ Server shutdown complete")
}

func JSONMarshal(v interface{}) ([]byte, error) {
	// Custom JSON marshaling with proper time formatting
	return json.Marshal(v)
}

func JSONUnmarshal(data []byte, v interface{}) error {
	// Custom JSON unmarshaling
	return json.Unmarshal(data, v)
}