package cache

import (
	"T-match_backend/internal/models"
	"context"
	"log"
	"os"
	"time"

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

type Redis struct {
	cache *redis.Client
}

func NewRedis(r *redis.Client) *Redis {
	return &Redis{cache: r}
}

func (r *Redis) Set(key string, value []byte, time time.Duration) error {
	err := r.cache.Set(context.Background(), key, value, time).Err()
	return err
}

func (r *Redis) Get(key string) (string, error) {
	value, err := r.cache.Get(context.Background(), key).Result()
	return value, err
}

func (r *Redis) Del(key string) {
	r.cache.Del(context.Background(), key).Result()
}
