package rw

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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

func RedisPing(config RedisConfig) {

}

func QueeryNewUser(user User, db *sql.DB) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, user.Email, user.PasswordHash, user.Role,
	).Scan(&id)
	return id, err
}
