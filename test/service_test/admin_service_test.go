// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetAdminStats(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT (SELECT COUNT(*) FROM interns) AS total_interns, (SELECT COUNT(*) FROM companies) AS total_companies, (SELECT COUNT(*) FROM internships WHERE is_archived = false) AS total_internships, (SELECT COUNT(*) FROM applications) AS total_responses`).
		WillReturnRows(sqlmock.NewRows([]string{"total_interns", "total_companies", "total_internships", "total_responses"}).
			AddRow(10, 5, 8, 25))

	env.mock.ExpectQuery(`SELECT COUNT(*) FILTER (WHERE status = 'pending') AS pending, COUNT(*) FILTER (WHERE status = 'reviewing') AS reviewing, COUNT(*) FILTER (WHERE status = 'accepted') AS accepted, COUNT(*) FILTER (WHERE status = 'rejected') AS rejected FROM applications`).
		WillReturnRows(sqlmock.NewRows([]string{"pending", "reviewing", "accepted", "rejected"}).
			AddRow(5, 3, 2, 1))

	env.mock.ExpectQuery(`SELECT (SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '7 days') AS new_users, (SELECT COUNT(*) FROM internships WHERE created_at > NOW() - INTERVAL '7 days') AS new_internships, (SELECT COUNT(*) FROM applications WHERE created_at > NOW() - INTERVAL '7 days') AS new_responses`).
		WillReturnRows(sqlmock.NewRows([]string{"new_users", "new_internships", "new_responses"}).
			AddRow(4, 3, 9))

	stats, err := env.svc.GetAdminStats(ctxWithClaims(1, constants.Admin, "admin@test.ru"))
	require.NoError(t, err)

	require.Equal(t, 10, stats.TotalInterns)
	require.Equal(t, 5, stats.TotalCompanies)
	require.Equal(t, 8, stats.TotalInternships)
	require.Equal(t, 25, stats.TotalResponses)
	require.Equal(t, 5, stats.ResponsesPending)
	require.Equal(t, 3, stats.ResponsesReviewing)
	require.Equal(t, 2, stats.ResponsesAccepted)
	require.Equal(t, 1, stats.ResponsesRejected)
	require.Equal(t, 4, stats.NewUsers7Days)
	require.Equal(t, 3, stats.NewInternships7Days)
	require.Equal(t, 9, stats.NewResponses7Days)
	require.Equal(t, 0, stats.UsersOnline)
}

func TestBanUserSelfBan(t *testing.T) {
	env := newTestService(t)

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.ru"), 5, models.AdminBanRequest{Reason: "spam"})
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestBanUserNoClaims(t *testing.T) {
	env := newTestService(t)

	err := env.svc.BanUser(context.Background(), 5, models.AdminBanRequest{Reason: "spam"})
	require.ErrorIs(t, err, apierrors.ErrInternalServer)
}

func TestBanUserValidationError(t *testing.T) {
	env := newTestService(t)

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4,
		models.AdminBanRequest{Reason: ""})
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestBanUserCannotBanAdmin(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(constants.Admin))

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4,
		models.AdminBanRequest{Reason: "spam"})
	require.ErrorIs(t, err, apierrors.ErrCannotBanAdmin)
}

func TestBanUserGetUserRoleError(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnError(errors.New("db down"))

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4,
		models.AdminBanRequest{Reason: "spam"})
	require.Error(t, err)
}

func TestBanUserAlreadyBanned(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(constants.Intern))

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id = $1)`).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4,
		models.AdminBanRequest{Reason: "spam"})
	require.ErrorIs(t, err, apierrors.ErrUserAlreadyBanned)
}

func TestBanUserSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(constants.Intern))

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id = $1)`).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	env.mock.ExpectExec(`INSERT INTO user_bans (user_id, reason, banned_by) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO UPDATE SET reason = $2, banned_by = $3, banned_at = NOW()`).
		WithArgs(4, "spam", 5).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := env.svc.BanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4,
		models.AdminBanRequest{Reason: "spam"})
	require.NoError(t, err)
}

func TestUnbanUserSelf(t *testing.T) {
	env := newTestService(t)

	err := env.svc.UnbanUser(ctxWithClaims(5, constants.Admin, "admin@test.ru"), 5)
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestUnbanUserNotFound(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnError(sql.ErrNoRows)

	err := env.svc.UnbanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4)
	require.ErrorIs(t, err, apierrors.ErrUserNotFound)
}

func TestUnbanUserNotBanned(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(constants.Intern))

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id = $1)`).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := env.svc.UnbanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4)
	require.ErrorIs(t, err, apierrors.ErrUserNotBanned)
}

func TestUnbanUserSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT role FROM users WHERE id = $1").
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(constants.Intern))

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id = $1)`).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	env.mock.ExpectExec(`DELETE FROM user_bans WHERE user_id = $1`).
		WithArgs(4).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := env.svc.UnbanUser(ctxWithClaims(5, constants.Admin, "admin@test.com"), 4)
	require.NoError(t, err)
}

func TestAdminDeleteInternshipNotFound(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM internships WHERE id = $1)`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := env.svc.AdminDeleteInternship(ctxWithClaims(1, constants.Admin, "admin@test.ru"), 42)
	require.ErrorIs(t, err, apierrors.ErrInternshipNotFound)
}

func TestAdminDeleteInternshipSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM internships WHERE id = $1)`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	env.mock.ExpectExec(`DELETE FROM internships WHERE id = $1`).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := env.svc.AdminDeleteInternship(ctxWithClaims(1, constants.Admin, "admin@test.ru"), 42)
	require.NoError(t, err)
}

func TestAdminDeleteInternshipExistsError(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM internships WHERE id = $1)`).
		WithArgs(42).
		WillReturnError(errors.New("db error"))

	err := env.svc.AdminDeleteInternship(ctxWithClaims(1, constants.Admin, "admin@test.ru"), 42)
	require.Error(t, err)
}

func TestAdminDeleteInternshipDeleteError(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM internships WHERE id = $1)`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	env.mock.ExpectExec(`DELETE FROM internships WHERE id = $1`).
		WithArgs(42).
		WillReturnError(errors.New("db error"))

	err := env.svc.AdminDeleteInternship(ctxWithClaims(1, constants.Admin, "admin@test.ru"), 42)
	require.Error(t, err)
}
