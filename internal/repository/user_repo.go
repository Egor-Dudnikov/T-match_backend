// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
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

	return db, nil
}

type Repository struct {
	db *sql.DB
}

func NewRepository(r *sql.DB) *Repository {
	return &Repository{db: r}
}

func (r *Repository) QueryNewUser(ctx context.Context, user models.User, birthDate time.Time) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, user.Email, user.PasswordHash, user.Role,
	).Scan(&id)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO interns (user_id, birth_date)
		VALUES ($1, $2)`, id, birthDate).Err()

	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	tx.Commit()
	return id, nil
}

func (r *Repository) QueryNewCompany(ctx context.Context, company models.User, companyData models.CompanyData) (int, error) {
	var id int
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, company.Email, company.PasswordHash, company.Role,
	).Scan(&id)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO companies (user_id, company_name, inn, kpp, ogrn, legal_address, director_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, id, companyData.ShortName, companyData.Inn, companyData.Kpp, companyData.Ogrn, companyData.Address, companyData.Director).Err()
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	tx.Commit()
	return id, nil
}

func (r *Repository) CheckUserEmail(ctx context.Context, email string, role string) (bool, error) {
	var exist bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2)`, email, role).Scan(&exist)
	if err != nil {
		return exist, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return exist, nil
}

func (r *Repository) GetUser(ctx context.Context, email string, role string) (models.User, error) {
	user := models.User{}
	err := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, role FROM users WHERE email = $1 AND role = $2`, email, role).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.Role)
	if err != nil {
		return user, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return user, nil
}

func (r *Repository) GetCompanyIdByUserId(ctx context.Context, userID int) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM companies WHERE user_id = $1`, userID).Scan(&id)
	return id, apierrors.Warp(apierrors.ErrDatabaseError, err)
}
