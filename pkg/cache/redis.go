package cache

import (
	"context"
	"fmt"
	"sync"

	config "github.com/Vighnesh-V-H/async/configs"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
    redisOnce sync.Once
    redisClient *redis.Client
    redisInitErr error
    redisLog *zerolog.Logger
)

func InitRedis(ctx context.Context, cfg *config.RedisConfig, log zerolog.Logger) (*redis.Client, error) {
    redisLog = &log
    redisOnce.Do(func() {
        redisLog.Info().
            Str("addr", cfg.Address).
            Int("db", cfg.DB).
            Msg("Initializing Redis connection")
        
        rdb := redis.NewClient(&redis.Options{
            Addr:         cfg.Address,
            Password:     cfg.Password,
            DB:           cfg.DB,
            PoolSize:     cfg.PoolSize,
            MinIdleConns: cfg.MinIdleConns,
            DialTimeout:  cfg.DialTimeout,
            ReadTimeout:  cfg.ReadTimeout,
            WriteTimeout: cfg.WriteTimeout,
        })

        redisLog.Info().
            Int("pool_size", cfg.PoolSize).
            Int("min_idle_conns", cfg.MinIdleConns).
            Dur("dial_timeout", cfg.DialTimeout).
            Msg("Redis connection pool configured")

        if err := rdb.Ping(ctx).Err(); err != nil {
            redisInitErr = fmt.Errorf("redis ping failed: %w", err)
            redisLog.Error().Err(err).Msg("Failed to ping Redis")
            return
        }

        redisClient = rdb
        redisLog.Info().Msg("Redis initialized and connected successfully")
    })

    if redisInitErr != nil {
        return nil, redisInitErr
    }
    return redisClient, nil
}

func GetRedis(cfg *config.RedisConfig, log zerolog.Logger) *redis.Client {
    client, err := InitRedis(context.Background(), cfg, log)
    if err != nil {
        panic(fmt.Sprintf("Failed to init Redis: %v", err))
    }
    return client
}

func CloseRedis() error {
    if redisClient == nil {
        return nil
    }
    if err := redisClient.Close(); err != nil {
        return fmt.Errorf("failed to close Redis: %w", err)
    }
    redisClient = nil
    redisLog.Info().Msg("Redis connection pool closed")
    return nil
}
