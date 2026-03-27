package cash

import (
	"T-match_backend/internal/models"
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func PingRedis(cfg models.RedisConfig) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	log.Println("Successfully connected to redis")
	return db, nil
}
