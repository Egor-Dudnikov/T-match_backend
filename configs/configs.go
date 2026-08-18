// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package configs

import (
	"T-match_backend/internal/models"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadConfig() models.Config {
	return models.Config{
		DbConfig: models.DbConfig{
			Host:    getEnv("DB_HOST", "postgres"),
			Port:    getEnvAsInt("DB_PORT", 5432),
			Name:    getEnv("DB_NAME", "t_match_database"),
			User:    getEnv("DB_USER", "realworld"),
			Sslmode: getEnv("DB_SSLMODE", "disable"),
		},
		ServerConfig: models.ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", ":8080"),
		},
		RedisConfig: models.RedisConfig{
			Addr:        getEnv("REDIS_ADDR", "redis:6379"),
			DB:          getEnvAsInt("REDIS_DB", 0),
			MaxRetries:  getEnvAsInt("REDIS_MAX_RETRIES", 3),
			DialTimeout: getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			Timeout:     getEnvAsDuration("REDIS_TIMEOUT", 3*time.Second),
		},
		EmailConfig: models.EmailConfig{
			Addr:     getEnv("EMAIL_ADDR", ""),
			Host:     getEnv("EMAIL_HOST", ""),
			Identity: getEnv("EMAIL_IDENTITY", ""),
			Username: getEnv("EMAIL_USERNAME", ""),
		},
		CORSConfig: models.CORSConfig{
			ControlAllowOrigin:  getEnv("CORS_ALLOW_ORIGIN", "http://localhost:8000"),
			ControlAllowHeaders: getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Content-Type", "Authorization", "X-Verify-Session", "X-New-Access-Token"}),
		},
		S3Config: models.S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", "minioadmin"),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", "minioadmin"),
			UseSSL:          getEnvAsBool("S3_USE_SSL", false),
		},
		RecsysConfig: models.RecsysConfig{
			URL: getEnv("RECSYS_URL", "http://recsys:8000"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(intValue)
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return defaultValue
}
