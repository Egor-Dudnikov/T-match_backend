package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id           int
	Email        string
	Role         string
	PasswordHash string
}

type UserRegistration struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
}

type UserVerify struct {
	Email        string
	PasswordHash string
	Code         string
}

type DbConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	User    string `json:"user"`
	Sslmode string `json:"sslmode"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RedisConfig struct {
	Addr        string        `json:"addr"`
	DB          int           `json:"db"`
	Max_retries int           `json:"max_retries"`
	DialTimeout time.Duration `json:"dial_timeout"`
	Timeout     time.Duration `json:"time_duration"`
}

type EmailConfig struct {
	Addr     string `json:"addr"`
	Host     string `json:"host"`
	Identity string `json:"identity"`
	Username string `json:"username"`
}

type VerifyRequest struct {
	Code string `json:"code"`
}

type Config struct {
	DbConfig     DbConfig
	ServerConfig ServerConfig
	RedisConfig  RedisConfig
	EmailConfig  EmailConfig
}

type Claims struct {
	UserID   string
	DeviceID string
	Email    string
	jwt.RegisteredClaims
}
