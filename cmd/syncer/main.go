package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/database"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/elasticsearch"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/redis"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	db := database.GetDB()

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Warning: Redis not available, sync service will use direct writes: %v", err)
	}

	esClient, err := elasticsearch.NewESClient(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobRepo := repository.NewJobRepository(db)
	userRepo := repository.NewUserRepository(db)

	jobIndexer := elasticsearch.NewJobIndexer(esClient, jobRepo, userRepo)
	candidateIndexer := elasticsearch.NewCandidateIndexer(esClient, userRepo)
	companyIndexer := elasticsearch.NewCompanyIndexer(esClient, userRepo)

	syncService := elasticsearch.NewSyncService(esClient, jobIndexer, candidateIndexer, companyIndexer)
	syncService.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down sync service...")
	syncService.Stop()

	if redisClient != nil {
		redisClient.Close()
	}

	log.Println("Sync service stopped")
}
