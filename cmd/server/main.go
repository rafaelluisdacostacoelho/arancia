package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arancia/todo-service/internal/config"
	"github.com/arancia/todo-service/internal/httpserver"
	"github.com/arancia/todo-service/internal/todo"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository based on configuration
	var repo todo.Repository
	switch cfg.StorageBackend {
	case "boltdb":
		repo, err = todo.NewBoltDBRepository(cfg.BoltDBPath)
		if err != nil {
			log.Fatalf("Failed to initialize BoltDB repository: %v", err)
		}
		defer func() {
			if err := repo.Close(); err != nil {
				log.Printf("Error closing repository: %v", err)
			}
		}()
	case "memory":
		repo = todo.NewMemoryRepository()
	default:
		log.Fatalf("Unknown storage backend: %s", cfg.StorageBackend)
	}

	// Create and start HTTP server
	server := httpserver.NewServer(cfg.Port, repo)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Port)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

