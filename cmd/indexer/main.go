package main

import (
	"context"
	"log"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/database"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/elasticsearch"
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

	// Initialize Elasticsearch
	esClient, err := elasticsearch.NewESClient(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	// Create indexes
	ctx := context.Background()
	if err := esClient.CreateIndex(ctx, "jobs", elasticsearch.JobIndexMapping); err != nil {
		log.Printf("Warning: Could not create jobs index: %v", err)
	}
	if err := esClient.CreateIndex(ctx, "candidates", elasticsearch.CandidateIndexMapping); err != nil {
		log.Printf("Warning: Could not create candidates index: %v", err)
	}
	if err := esClient.CreateIndex(ctx, "companies", elasticsearch.CompanyIndexMapping); err != nil {
		log.Printf("Warning: Could not create companies index: %v", err)
	}

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize indexer
	jobIndexer := elasticsearch.NewJobIndexer(esClient, jobRepo, userRepo)

	// Index all jobs
	log.Println("Starting full reindex...")
	startTime := time.Now()

	if err := jobIndexer.IndexAllJobs(ctx); err != nil {
		log.Printf("Error indexing jobs: %v", err)
	}

	log.Printf("Reindex completed in %v", time.Since(startTime))
}