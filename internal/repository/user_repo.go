// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"fmt"
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

	return db, nil
}

type Repository struct {
	db *sql.DB
}

func NewRepository(r *sql.DB) *Repository {
	return &Repository{db: r}
}

func (r *Repository) QueryNewUser(ctx context.Context, user models.InternVerify) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, user.Email, user.PasswordHash, constants.Intern,
	).Scan(&id)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO interns (user_id, birth_date)
		VALUES ($1, $2)`, id, user.BirthDate).Err()

	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}

	return id, nil
}

func (r *Repository) QueryNewCompany(ctx context.Context, company models.CompanyVerify) (int, error) {
	var id int
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, role, created_at)
        VALUES ($1, $2, $3, NOW())
		RETURNING id`, company.Email, company.PasswordHash, constants.Company,
	).Scan(&id)
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}

	err = tx.QueryRowContext(ctx, `INSERT INTO companies (user_id, company_name, inn, kpp, ogrn, legal_address, director_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, id, company.CompanyData.ShortName,
		company.CompanyData.Inn,
		company.CompanyData.Kpp,
		company.CompanyData.Ogrn,
		company.CompanyData.Address,
		company.CompanyData.Director).Err()

	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}

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
		&user.ID,
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
	if err != nil {
		return id, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return id, nil
}

func (r *Repository) GetEmailByUserId(ctx context.Context, id int) (string, error) {
	var email string
	err := r.db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1`, id).Scan(&email)
	if err != nil {
		return email, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return email, err
}

func (r *Repository) GetUserIdByCompanyId(ctx context.Context, id int) (int, error) {
	var userID int
	err := r.db.QueryRowContext(ctx, `SELECT user_id FROM companies WHERE id = $1`, id).Scan(&userID)
	if err != nil {
		return userID, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return userID, err
}

func (r *Repository) GetUserIdByEmail(ctx context.Context, email string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err != nil {
		return id, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return id, err
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, newPasswordHash string, id int) error {
	err := r.db.QueryRowContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, newPasswordHash, id).Err()
	if err != nil {
		return apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return nil
}
