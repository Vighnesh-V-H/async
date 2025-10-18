package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var (
    redisOnce sync.Once
    redisClient *redis.Client
    redisInitErr error
    redisLog *zerolog.Logger
)

func InitRedis(ctx context.Context, addr, password string, db int, log zerolog.Logger) (*redis.Client, error) {
    redisLog = &log
    redisOnce.Do(func() {
        redisLog.Info().
            Str("addr", addr).
            Int("db", db).
            Msg("Initializing Redis connection")
        
        rdb := redis.NewClient(&redis.Options{
            Addr:         addr,
            Password:     password,
            DB:           db,
            PoolSize:     20,
            MinIdleConns: 5,
            DialTimeout:  5 * time.Second,
            ReadTimeout:  2 * time.Second,
            WriteTimeout: 2 * time.Second,
        })

        redisLog.Info().
            Int("pool_size", 20).
            Int("min_idle_conns", 5).
            Dur("dial_timeout", 5*time.Second).
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

func GetRedis(addr, password string, db int, log zerolog.Logger) *redis.Client {
    client, err := InitRedis(context.Background(), addr, password, db, log)
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
