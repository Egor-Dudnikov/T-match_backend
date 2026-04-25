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

func (r *Redis) RateLimitCheck(ctx context.Context, key string, rate int) (bool, error) {
	now := time.Now().Unix()

	rate_script := `local key = KEYS[1]
					local rate = tonumber(ARGV[1])
					local time_now = tonumber(ARGV[2])

					local exits = redis.call('EXISTS', key)

					if exits == 0 then 
						redis.call('HSET', key,  "token", rate)
						redis.call('HSET', key, "last_time", time_now)
					end

					local token = tonumber(redis.call('HGET', key, "token"))
					local last_time = tonumber(redis.call('HGET', key, "last_time"))

					local time = time_now - last_time

					local limit = 60 / rate

					local cnt_limit = math.floor(time / limit)

					if cnt_limit + token > rate then
						token = rate
					else
						token = token + cnt_limit
					end

					if token == 0 then 
						redis.call('EXPIRE', key, 120)
						return 0
					else 
						token = token - 1
						redis.call('HSET', key, "token", token)
						redis.call('HSET', key, "last_time", time_now)
						redis.call('EXPIRE', key, 120)
						return 1
					end`

	cmd := r.cache.Eval(ctx, rate_script, []string{key}, rate, now)
	res, err := cmd.Int64()
	if err != nil {
		return false, err
	}
	if res == 1 {
		return true, nil
	}
	return false, nil
}
