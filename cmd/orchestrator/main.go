package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/Vighnesh-V-H/async/configs"
	"github.com/Vighnesh-V-H/async/internal/events"
	"github.com/Vighnesh-V-H/async/internal/handler"
	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/Vighnesh-V-H/async/internal/orchestrator"
	"github.com/Vighnesh-V-H/async/internal/repositories"
	"github.com/Vighnesh-V-H/async/internal/router"
	"github.com/Vighnesh-V-H/async/internal/service"
	"github.com/Vighnesh-V-H/async/pkg/database"
	"github.com/Vighnesh-V-H/async/pkg/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger from config
	logCfg := logger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		ServiceName: cfg.Logger.ServiceName,
		Environment: cfg.Logger.Environment,
		IsProd:      cfg.Logger.IsProd,
	}

	appLog := logger.New(logCfg)

	appLog.Info().
		Str("environment", cfg.Primary.Environment).
		Str("service", cfg.Primary.ServiceName).
		Msg("Starting orchestrator service")

	// Initialize database with config
	appLog.Info().Msg("Initializing database connection")
	db, err := database.InitDB(ctx, &cfg.Database, logCfg)
	if err != nil {
		appLog.Fatal().Err(err).Msg("Failed to initialize database")
	}

	defer func() {
		if err := database.CloseDB(ctx); err != nil {
			appLog.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	appLog.Info().Msg("Database initialized successfully")

	// Initialize Kafka producer with config
	appLog.Info().Msg("Initializing Kafka producer")
	kafkaProducer, err := kafka.InitProducer(&cfg.Kafka, appLog)
	if err != nil {
		appLog.Fatal().Err(err).Msg("Failed to initialize Kafka producer")
	}
	defer kafka.Close()

	appLog.Info().Msg("Kafka producer initialized successfully")

	// Initialize Kafka consumer with config
	appLog.Info().Msg("Initializing Kafka consumer for task completions")
	kafkaConsumer, err := kafka.InitConsumer(&cfg.Kafka, appLog)
	if err != nil {
		appLog.Fatal().Err(err).Msg("Failed to initialize Kafka consumer")
	}

	appLog.Info().Msg("Kafka consumer initialized successfully")

	// Initialize event handlers
	eventProducer := events.NewEventProducer(kafkaProducer, logCfg)
	eventConsumer := events.NewEventConsumer(kafkaConsumer, logCfg)
	appLog.Info().Msg("Event producer and consumer initialized")

	// Initialize repositories
	workflowRepo := repositories.NewWorkflowRepository(db)
	instanceRepo := repositories.NewInstanceRepository(db)
	appLog.Info().Msg("Repositories initialized")

	// Initialize services
	workflowService := service.NewWorkflowService(workflowRepo)
	instanceService := service.NewInstanceService(instanceRepo)
	appLog.Info().Msg("Services initialized")

	// Initialize orchestrator
	orch := orchestrator.NewOrchestrator(workflowService, instanceService, eventProducer, logCfg)
	appLog.Info().Msg("Orchestrator state machine initialized")

	// Initialize handlers
	workflowHandler := handler.NewWorkflowHandler(workflowService)
	audioHandler := handler.NewAudioHandler(workflowService, instanceService, eventProducer)
	appLog.Info().Msg("Handlers initialized")

	// Start Kafka consumer in background
	go func() {
		appLog.Info().Msg("Starting Kafka consumer for task completions")
		if err := eventConsumer.ConsumeCompletions(ctx, orch.ProcessCompletion); err != nil {
			appLog.Error().Err(err).Msg("Kafka consumer stopped")
		}
	}()

	// Setup Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	
	ginRouter := gin.Default()

	ginRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     cfg.Primary.ServiceName,
			"environment": cfg.Primary.Environment,
		})
	})

	router.SetupWorkflowRoutes(ginRouter, workflowHandler)
	router.SetupAudioRoutes(ginRouter, audioHandler)
	appLog.Info().Msg("Routes configured")

	// Setup HTTP server with config
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start HTTP server in background
	go func() {
		appLog.Info().Str("port", cfg.Server.Port).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	appLog.Info().Msg("Orchestrator started, waiting for shutdown signal")
	<-sigChan

	appLog.Info().Msg("Shutting down gracefully...")

	// Shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error().Err(err).Msg("Server forced to shutdown")
	}

	cancel()
	appLog.Info().Msg("Orchestrator stopped")
}
