package repository

import (
	"T-match_backend/internal/models"
	"database/sql"
	"fmt"
	"log"
	"os"
)

func PingDatabase(config models.DbConfig) (*sql.DB, error) {

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

func QueeryNewUser(user models.User, db *sql.DB) (int, error) {
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
