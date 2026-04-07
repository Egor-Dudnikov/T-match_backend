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

type Repository struct {
	db *sql.DB
}

func NewRepository(r *sql.DB) *Repository {
	return &Repository{db: r}
}

func (r *Repository) QueryNewUser(user models.User) (int, error) {
	var id int
	err := r.db.QueryRow(`INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, user.Email, user.PasswordHash, user.Role,
	).Scan(&id)
	return id, err
}

func (r *Repository) CheckUserEmail(email string) (bool, error) {
	var exist bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exist)
	if err != nil {
		return exist, err
	}
	return exist, nil
}

func (r *Repository) GetUser(email string) (models.User, error) {
	user := models.User{}
	err := r.db.QueryRow(`SELECT id, email, password_hash, role FROM users WHERE email = $1`, email).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}
