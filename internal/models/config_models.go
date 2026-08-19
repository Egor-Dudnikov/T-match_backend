// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import (
	"time"
)

// Config is the root application configuration.
type Config struct {
	DbConfig     DbConfig
	ServerConfig ServerConfig
	RedisConfig  RedisConfig
	EmailConfig  EmailConfig
	CORSConfig   CORSConfig
	S3Config     S3Config
	RecsysConfig RecsysConfig
	CookieSecure bool
}

// DbConfig holds the PostgreSQL connection settings.
type DbConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	User    string `json:"user"`
	Sslmode string `json:"sslmode"`
}

// ServerConfig holds the HTTP server listen settings.
type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// RedisConfig holds the Redis connection settings.
type RedisConfig struct {
	Addr        string        `json:"addr"`
	DB          int           `json:"db"`
	MaxRetries  int           `json:"max_retries"`
	DialTimeout time.Duration `json:"dial_timeout"`
	Timeout     time.Duration `json:"time_duration"`
}

// EmailConfig holds the SMTP server settings.
type EmailConfig struct {
	Addr     string `json:"addr"`
	Host     string `json:"host"`
	Identity string `json:"identity"`
	Username string `json:"username"`
}

// CORSConfig holds the CORS allowlist settings.
type CORSConfig struct {
	ControlAllowOrigin  string   `json:"control_allow_origin"`
	ControlAllowHeaders []string `json:"control_allow_headers"`
}

// S3Config holds the S3-compatible storage credentials.
type S3Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_ID"`
	SecretAccessKey string `json:"secret_access_key"`
	UseSSL          bool   `json:"use_SSL"`
}

// RecsysConfig holds the recommendation service connection settings.
type RecsysConfig struct {
	URL string `json:"url"`
}
