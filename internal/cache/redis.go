// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

// Package cache provides the Redis-backed session and rate-limit store.
package cache

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// PingRedis creates a Redis client from the given config and verifies the connection.
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

// Redis is a cache wrapper around a Redis client.
type Redis struct {
	cache *redis.Client
}

// NewRedis wraps the given Redis client in a cache.
func NewRedis(r *redis.Client) *Redis {
	return &Redis{cache: r}
}

// Close closes the underlying Redis connection.
func (r *Redis) Close() error {
	return r.cache.Close()
}

// Set stores the given value under the key with the specified time-to-live.
func (r *Redis) Set(ctx context.Context, key string, value []byte, time time.Duration) error {
	err := r.cache.Set(ctx, key, value, time).Err()
	if err != nil {
		return apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	return nil
}

// Get returns the value stored under the given key.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return value, apierrors.ErrKeyNotFound
		}
		return value, apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	return value, nil
}

// Del removes the value stored under the given key.
func (r *Redis) Del(ctx context.Context, key string) error {
	_, err := r.cache.Del(ctx, key).Result()
	if err != nil {
		return apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	return nil
}

// DeleteUserSessions removes all cached keys belonging to the given user.
func (r *Redis) DeleteUserSessions(ctx context.Context, userID int) error {
	prefix := strconv.Itoa(userID) + "."

	var keys []string
	iter := r.cache.Scan(ctx, 0, prefix+"*", 1000).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		return apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	if len(keys) == 0 {
		return nil
	}

	_, err := r.cache.Del(ctx, keys...).Result()
	if err != nil {
		return apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	return nil
}

// RateLimitCheck reports whether the request for the given key is allowed under the rate limit.
func (r *Redis) RateLimitCheck(ctx context.Context, key string, rate int) (bool, error) {
	now := time.Now().Unix()

	rateScript := `local key = KEYS[1]
					local rate = tonumber(ARGV[1])
					local time_now = tonumber(ARGV[2])

					local exists = redis.call('EXISTS', key)

					if exists == 0 then 
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

	cmd := r.cache.Eval(ctx, rateScript, []string{key}, rate, now)
	res, err := cmd.Int64()
	if err != nil {
		return false, apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	if res == 1 {
		return true, nil
	}
	return false, nil
}

// ResetCode replaces the value of the given key while preserving its time-to-live.
func (r *Redis) ResetCode(ctx context.Context, key, newValue string) error {
	script := `local key = KEYS[1]
				local value = ARGV[1]

				local ttl = redis.call('TTL', KEYS[1])
				if ttl < 0 then 
					return 0
				end
				redis.call('DEL', key)

				redis.call('SET', key, value)
				redis.call('EXPIRE', key, ttl)
				return 1`
	cmd := r.cache.Eval(ctx, script, []string{key}, newValue)
	ok, err := cmd.Int64()
	if err != nil {
		return apierrors.Wrap(apierrors.ErrCacheError, err)
	}
	if ok == 0 {
		return redis.Nil
	}
	return nil
}
