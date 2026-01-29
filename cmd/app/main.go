package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"familyshare/internal/config"
	"familyshare/internal/db"
	"familyshare/internal/handler"
	"familyshare/internal/janitor"
	"familyshare/internal/storage"
	"familyshare/web"
)

func main() {
	// Load config from environment
	cfg := config.Load()

	// Initialize database
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize storage
	store := storage.New(cfg.DataDir)

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Initialize handlers
	h := handler.New(database, store, web.EmbedFS, cfg)

	// Register routes
	h.RegisterRoutes(r)

	// Initialize and start janitor for cleanup tasks
	jan := janitor.New(janitor.Config{
		DB:          database,
		StoragePath: cfg.DataDir,
		TempUploadDir: cfg.TempUploadDir,
		Interval:    cfg.JanitorInterval,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jan.Start(ctx)
	defer jan.Stop()

	// Create server
	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("FamilyShare starting on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
