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
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
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

type VeryfyConfig struct {
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
	VeryfyConfig VeryfyConfig
}

type Claims struct {
	UserID   string
	DeviceID string
	Email    string
	jwt.RegisteredClaims
}
