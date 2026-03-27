package rw

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func PingDatabase(config DbConfig) (*sql.DB, error) {

	conStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, os.Getenv("DB_PASSWORD"), config.Name, config.Sslmode)

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")

	return db, nil
}

func PingRedis(cfg RedisConfig) (*redis.Client, error) {
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

func QueeryNewUser(user User, db *sql.DB) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, user.Email, user.PasswordHash, user.Role,
	).Scan(&id)
	return id, err
}

func CheckUserEmail(email string, db *sql.DB) (bool, error) {
	var exist bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exist)
	if err != nil {
		return exist, err
	}
	return exist, nil
}
