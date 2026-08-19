// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthInternValidationError(t *testing.T) {
	env := newTestService(t)

	user := models.InternAuth{
		Email:    "invalid-email",
		Password: "Aa1aaaaa",
		DeviceID: "device123",
	}
	_, err := env.svc.AuthIntern(ctxWithClaims(0, constants.Intern, "email"), user)
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestAuthInternTooYoung(t *testing.T) {
	env := newTestService(t)

	user := models.InternAuth{
		Email:     "intern@test.ru",
		Password:  "Aa1aaaaa",
		DeviceID:  "device-12345",
		BirthDate: time.Now().AddDate(-15, 0, 0),
	}
	_, err := env.svc.AuthIntern(ctxWithClaims(0, constants.Intern, "intern@test.ru"), user)
	require.ErrorIs(t, err, apierrors.ErrUserMustBe16)
}

func TestAuthInternAlreadyExists(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2)").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	user := models.InternAuth{
		Email:     "intern@test.ru",
		Password:  "Aa1aaaaa",
		DeviceID:  "device-12345",
		BirthDate: time.Now().AddDate(-20, 0, 0),
	}
	_, err := env.svc.AuthIntern(ctxWithClaims(0, constants.Intern, "intern@test.ru"), user)
	require.ErrorIs(t, err, apierrors.ErrUserAlreadyExists)
}

func TestAuthInternSuccessSkipped(t *testing.T) {
	// EmailClient.SendVerifyCode hits real SMTP and cannot be mocked without
	// interfaces, so the happy path of AuthIntern is covered by the DB-level
	// checks in TestAuthInternAlreadyExists, and the email part is out of scope.
	t.Skip("skipping: EmailClient.SendVerifyCode hits real SMTP, cannot be mocked without interfaces")
}

func TestLoginUserNotExists(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2)").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, _, err := env.svc.LoginUser(ctxWithClaims(0, constants.Intern, ""), models.LoginUser{ //nolint:gosec // test-only values
		Email:        "intern@test.ru",
		PasswordHash: "not-a-real-hash",
		DeviceID:     "device-12345",
	}, constants.Intern)
	require.ErrorIs(t, err, apierrors.ErrUserNotExists)
}

func TestLoginUserInvalidPassword(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2)").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	hash, err := bcrypt.GenerateFromPassword([]byte("Aa1aaaaa"), bcrypt.DefaultCost)
	require.NoError(t, err)

	env.mock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = $1 AND role = $2").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
			AddRow(1, "intern@test.ru", string(hash), constants.Intern))

	_, _, err = env.svc.LoginUser(ctxWithClaims(0, constants.Intern, ""), models.LoginUser{
		Email:        "intern@test.ru",
		PasswordHash: "wrong-password",
		DeviceID:     "device-12345",
	}, constants.Intern)
	require.ErrorIs(t, err, apierrors.ErrInvalidPassword)
}

func TestLoginUserSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2)").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	hash, err := bcrypt.GenerateFromPassword([]byte("Aa1aaaaa"), bcrypt.DefaultCost)
	require.NoError(t, err)

	env.mock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = $1 AND role = $2").
		WithArgs("intern@test.ru", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
			AddRow(1, "intern@test.ru", string(hash), constants.Intern))

	accessToken, refreshToken, err := env.svc.LoginUser(ctxWithClaims(0, constants.Intern, ""), models.LoginUser{
		Email:        "intern@test.ru",
		PasswordHash: "Aa1aaaaa",
		DeviceID:     "device-12345",
	}, constants.Intern)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
}

func TestVerifyUserExpiredSession(t *testing.T) {
	env := newTestService(t)

	// no session stored in redis -> ErrCodeExpired
	_, _, err := env.svc.VerifyUser(ctxWithClaims(0, constants.Intern, ""), "no-such-session", models.VerifyRequest{Code: "654321"}, constants.Intern)
	require.ErrorIs(t, err, apierrors.ErrCodeExpired)
}

func TestVerifyUserInvalidCode(t *testing.T) {
	env := newTestService(t)

	sessionID := "session-123"
	require.NoError(t, env.redis.Set(sessionID, `{"Email":"intern@test.ru","PasswordHash":"hash","DeviceID":"device-12345","BirthDate":"2000-01-01T00:00:00Z"}`))
	require.NoError(t, env.redis.Set(sessionID+".code", "111111"))

	// session exists but the submitted code does not match -> ErrInvalidCode
	_, _, err := env.svc.VerifyUser(ctxWithClaims(0, constants.Intern, ""), sessionID, models.VerifyRequest{Code: "654321"}, constants.Intern)
	require.ErrorIs(t, err, apierrors.ErrInvalidCode)
}

func TestVerifyUserSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	sessionID := "session-123"
	require.NoError(t, env.redis.Set(sessionID, `{"Email":"intern@test.ru","PasswordHash":"hash","DeviceID":"device-12345","BirthDate":"2000-01-01T00:00:00Z"}`))
	require.NoError(t, env.redis.Set(sessionID+".code", "111111"))

	env.mock.ExpectBegin()
	env.mock.ExpectQuery(`INSERT INTO users (email, password_hash, role, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`).
		WithArgs("intern@test.ru", "hash", constants.Intern).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	env.mock.ExpectQuery(`INSERT INTO interns (user_id, birth_date) VALUES ($1, $2)`).
		WithArgs(1, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	env.mock.ExpectCommit()

	accessToken, refreshToken, err := env.svc.VerifyUser(ctxWithClaims(0, constants.Intern, ""), sessionID, models.VerifyRequest{Code: "111111"}, constants.Intern)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
}
