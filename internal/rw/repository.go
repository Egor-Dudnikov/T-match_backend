package rw

import (
	"database/sql"
	"fmt"
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
		return db, nil
	}

	return db, nil
}

func QueeryNewUser(user User, db *sql.DB) int {
	var id int
	db.QueryRow(`INSERT INTO users ()
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id`,
	).Scan(&id)
	return id
}
