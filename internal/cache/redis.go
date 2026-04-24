package cache

import (
	"T-match_backend/internal/models"
	"context"
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
	return db, nil
}

type Redis struct {
	cache *redis.Client
}

func NewRedis(r *redis.Client) *Redis {
	return &Redis{cache: r}
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, time time.Duration) error {
	err := r.cache.Set(ctx, key, value, time).Err()
	return err
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.cache.Get(ctx, key).Result()
	return value, err
}

func (r *Redis) Del(ctx context.Context, key string) {
	r.cache.Del(ctx, key).Result()
}
