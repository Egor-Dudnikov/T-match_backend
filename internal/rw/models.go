package rw

import (
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
	Addr string `json:"addr"`
	DB   int    `json:"db"`
}

type Config struct {
	DbConfig     DbConfig
	ServerConfig ServerConfig
	RedisConfig  RedisConfig
}

type Claims struct {
	UserId string
	Email  string
	jwt.RegisteredClaims
}
