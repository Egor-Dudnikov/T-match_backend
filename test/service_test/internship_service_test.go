// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNewInternshipValidationError(t *testing.T) {
	env := newTestService(t)

	_, err := env.svc.NewInternship(ctxWithClaims(1, constants.Company, "company@test.ru"), models.Internship{
		Title:       "",
		Description: "",
		CityID:      0,
	}, 1)
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestNewInternshipSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT id FROM companies WHERE user_id = $1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	env.mock.ExpectQuery(`INSERT INTO internships (company_id, title, description, salary, duration_months, city_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id;`).
		WithArgs(2, "Go developer", "Backend internship", 100000, 6, 77).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	id, err := env.svc.NewInternship(ctxWithClaims(1, constants.Company, "company@test.ru"), models.Internship{
		Title:         "Go developer",
		Description:   "Backend internship",
		Salary:        100000,
		DurationMonth: 6,
		CityID:        77,
	}, 1)
	require.NoError(t, err)
	require.Equal(t, 7, id)
}

func TestGetInternshipByIDSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT i.id, i.company_id, i.title, i.description, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived FROM internships i JOIN companies comp ON i.company_id = comp.id WHERE i.id = $1 AND i.is_archived = FALSE AND NOT EXISTS(SELECT 1 FROM user_bans ub WHERE ub.user_id = comp.user_id)`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "company_id", "title", "description", "salary", "duration_months", "city_id", "created_at", "is_archived"}).
			AddRow(7, 2, "Go developer", "backend", 100000, 6, 77, time.Now(), false))

	env.mock.ExpectQuery(`SELECT skill_id FROM internship_skills WHERE internship_id = $1`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"skill_id"}))

	res, err := env.svc.GetInternshipByID(ctxWithClaims(1, constants.Intern, "intern@test.ru"), 7)
	require.NoError(t, err)
	require.Equal(t, 7, res.Internship.ID)
	require.Equal(t, "Go developer", res.Internship.Title)
	require.Equal(t, 77, res.Internship.CityID)
	require.Empty(t, res.Skills)
}

func TestGetInternshipByIDNotFound(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT i.id, i.company_id, i.title, i.description, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived FROM internships i JOIN companies comp ON i.company_id = comp.id WHERE i.id = $1 AND i.is_archived = FALSE AND NOT EXISTS(SELECT 1 FROM user_bans ub WHERE ub.user_id = comp.user_id)`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	_, err := env.svc.GetInternshipByID(ctxWithClaims(1, constants.Intern, "intern@test.ru"), 999)
	require.ErrorIs(t, err, apierrors.ErrInternshipNotFound)
}

func TestRespondNoClaims(t *testing.T) {
	env := newTestService(t)

	err := env.svc.RespondInternship(context.Background(), 7)
	require.ErrorIs(t, err, apierrors.ErrInternalServer)
}

func TestGetCompanyesInternshipsByUserID(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT id FROM companies WHERE user_id = $1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	env.mock.ExpectQuery(`SELECT i.id, i.company_id, i.title, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived FROM internships i WHERE i.company_id = $1`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "company_id", "title", "salary", "duration_months", "city_id", "created_at", "is_archived"}).
			AddRow(7, 2, "Go developer", 100000, 6, 77, time.Now(), false))

	res, err := env.svc.GetCompanyesInternshipsByUserID(context.TODO(), 1)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Go developer", res[0].Title)
}
