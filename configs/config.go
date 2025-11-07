package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)


type Config struct {
	Primary  PrimaryConfig  `koanf:"primary" validate:"required"`
	Server   ServerConfig   `koanf:"server" validate:"required"`
	Database DatabaseConfig `koanf:"database" validate:"required"`
	Redis    RedisConfig    `koanf:"redis" validate:"required"`
	Kafka    KafkaConfig    `koanf:"kafka" validate:"required"`
	Logger   LoggerConfig   `koanf:"logger" validate:"required"`
}

type PrimaryConfig struct {
	Environment string `koanf:"environment" validate:"required,oneof=development staging production"`
	ServiceName string `koanf:"service_name" validate:"required"`
}

type ServerConfig struct {
	Port         string        `koanf:"port" validate:"required"`
	ReadTimeout  time.Duration `koanf:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `koanf:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `koanf:"idle_timeout" validate:"required"`
}

type DatabaseConfig struct {
	URL             string        `koanf:"url" validate:"required"`
	MaxOpenConns    int           `koanf:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int           `koanf:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime time.Duration `koanf:"conn_max_idle_time" validate:"required"`
}

type RedisConfig struct {
	Address      string        `koanf:"address" validate:"required"`
	Password     string        `koanf:"password"`
	DB           int           `koanf:"db"`
	PoolSize     int           `koanf:"pool_size" validate:"required,min=1"`
	MinIdleConns int           `koanf:"min_idle_conns" validate:"required,min=1"`
	DialTimeout  time.Duration `koanf:"dial_timeout" validate:"required"`
	ReadTimeout  time.Duration `koanf:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `koanf:"write_timeout" validate:"required"`
}

type KafkaConfig struct {
	Brokers            []string      `koanf:"brokers" validate:"required,min=1"`
	ProducerAcks       string        `koanf:"producer_acks" validate:"required"`
	ProducerRetries    int           `koanf:"producer_retries" validate:"required,min=0"`
	ProducerLingerMs   int           `koanf:"producer_linger_ms" validate:"required,min=0"`
	CompressionType    string        `koanf:"compression_type" validate:"required"`
	RequestTimeoutMs   int           `koanf:"request_timeout_ms" validate:"required,min=1000"`
	DeliveryTimeoutMs  int           `koanf:"delivery_timeout_ms" validate:"required,min=1000"`
	ConsumerGroupID    string        `koanf:"consumer_group_id" validate:"required"`
	ConsumerTopic      string        `koanf:"consumer_topic" validate:"required"`
	AutoOffsetReset    string        `koanf:"auto_offset_reset" validate:"required,oneof=earliest latest none"`
	EnableAutoCommit   bool          `koanf:"enable_auto_commit"`
	SessionTimeoutMs   int           `koanf:"session_timeout_ms" validate:"required,min=1000"`
	HeartbeatInterval  time.Duration `koanf:"heartbeat_interval" validate:"required"`
}

type LoggerConfig struct {
	Level       string `koanf:"level" validate:"required,oneof=debug info warn error"`
	Format      string `koanf:"format" validate:"required,oneof=json console"`
	ServiceName string `koanf:"service_name"`
	Environment string `koanf:"environment"`
	IsProd      bool   `koanf:"is_prod"`
}

func LoadConfig() (*Config, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New(".")

	err := k.Load(env.Provider("ASYNC_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "ASYNC_"))
	}), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load environment variables")
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal configuration")
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	applyDefaults(mainConfig)

	validate := validator.New()
	err = validate.Struct(mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("configuration validation failed")
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	mainConfig.Logger.ServiceName = mainConfig.Primary.ServiceName
	mainConfig.Logger.Environment = mainConfig.Primary.Environment
	mainConfig.Logger.IsProd = mainConfig.Primary.Environment == "production"

	logger.Info().
		Str("environment", mainConfig.Primary.Environment).
		Str("service", mainConfig.Primary.ServiceName).
		Msg("Configuration loaded successfully")

	return mainConfig, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 15 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 15 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 60 * time.Second
	}

	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.Database.ConnMaxIdleTime == 0 {
		cfg.Database.ConnMaxIdleTime = 5 * time.Minute
	}

	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 20
	}
	if cfg.Redis.MinIdleConns == 0 {
		cfg.Redis.MinIdleConns = 5
	}
	if cfg.Redis.DialTimeout == 0 {
		cfg.Redis.DialTimeout = 5 * time.Second
	}
	if cfg.Redis.ReadTimeout == 0 {
		cfg.Redis.ReadTimeout = 2 * time.Second
	}
	if cfg.Redis.WriteTimeout == 0 {
		cfg.Redis.WriteTimeout = 2 * time.Second
	}

	if cfg.Kafka.ProducerAcks == "" {
		cfg.Kafka.ProducerAcks = "all"
	}
	if cfg.Kafka.ProducerRetries == 0 {
		cfg.Kafka.ProducerRetries = 3
	}
	if cfg.Kafka.ProducerLingerMs == 0 {
		cfg.Kafka.ProducerLingerMs = 5
	}
	if cfg.Kafka.CompressionType == "" {
		cfg.Kafka.CompressionType = "gzip"
	}
	if cfg.Kafka.RequestTimeoutMs == 0 {
		cfg.Kafka.RequestTimeoutMs = 30000
	}
	if cfg.Kafka.DeliveryTimeoutMs == 0 {
		cfg.Kafka.DeliveryTimeoutMs = 120000
	}
	if cfg.Kafka.AutoOffsetReset == "" {
		cfg.Kafka.AutoOffsetReset = "earliest"
	}
	if cfg.Kafka.SessionTimeoutMs == 0 {
		cfg.Kafka.SessionTimeoutMs = 6000
	}
	if cfg.Kafka.HeartbeatInterval == 0 {
		cfg.Kafka.HeartbeatInterval = 3 * time.Second
	}

	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "console"
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Primary.Environment == "development"
}


func (c *Config) IsProduction() bool {
	return c.Primary.Environment == "production"
}

func (c *Config) IsStaging() bool {
	return c.Primary.Environment == "staging"
}
