package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/Vighnesh-V-H/async/pkg/database"
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

	// dsn := fmt.Sprintf(
	// 	"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable  Timezone=UTC",
	// 	getEnv("DB_HOST", "localhost"),
	// 	getEnv("DB_USER", "pg"),
	// 	getEnv("DB_PASSWORD", "pradyuman"),
	// 	getEnv("DB_NAME", "db"),
	// 	getEnv("DB_PORT", "5432"),

	// )


	dbURL :=os.Getenv("DATABASE_URL")
	if dbURL==""{
		log.Error().Msg("Failed to get database url")
	}
		
	log.Info().Msg("Initializing database connection")

	_, err := database.InitDB(ctx, dbURL, logCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	

	defer func() {
		if err := database.CloseDB(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	log.Info().Msg("Database initialized successfully")


	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Info().Msg("Orchestrator started, waiting for shutdown signal")
	<-sigChan

	log.Info().Msg("Shutting down gracefully...")
	cancel()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
