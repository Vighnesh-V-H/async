package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vighnesh-V-H/async/internal/handler"
	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/Vighnesh-V-H/async/internal/repositories"
	"github.com/Vighnesh-V-H/async/internal/router"
	"github.com/Vighnesh-V-H/async/internal/service"
	"github.com/Vighnesh-V-H/async/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

  erro := godotenv.Load(".env")
  if erro != nil {
    log.Fatal("Error loading .env file")
  }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logCfg := logger.Config{
		Level:   getEnv("LOG_LEVEL", "info"),
		Format:  getEnv("LOG_FORMAT", "json"),
	}

	log := logger.New(logCfg)

	dbURL :=os.Getenv("DATABASE_URL")
	if dbURL==""{
		log.Error().Msg("Failed to get database url")
	}
		
	log.Info().Msg("Initializing database connection")

	db, err := database.InitDB(ctx, dbURL, logCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	defer func() {
		if err := database.CloseDB(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	log.Info().Msg("Database initialized successfully")

	workflowRepo := repositories.NewWorkflowRepository(db)
	log.Info().Msg("Workflow repository initialized")

	workflowService := service.NewWorkflowService(workflowRepo)
	log.Info().Msg("Workflow service initialized")

	workflowHandler := handler.NewWorkflowHandler(workflowService)
	log.Info().Msg("Workflow handler initialized")

	ginRouter := gin.Default()
	
	ginRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "orchestrator",
		})
	})

	router.SetupWorkflowRoutes(ginRouter, workflowHandler)
	log.Info().Msg("Workflow routes configured")

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: ginRouter,
	}

	go func() {
		log.Info().Str("port", port).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Info().Msg("Orchestrator started, waiting for shutdown signal")
	<-sigChan

	log.Info().Msg("Shutting down gracefully...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	cancel()
	log.Info().Msg("Orchestrator stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
