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

func TestUpdateStudentProfileValidationError(t *testing.T) {
	env := newTestService(t)

	tooLong := string(make([]byte, 101))
	profile := models.Profile{
		FirstName: &tooLong,
	}
	err := env.svc.UpdateStudentProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"), profile)
	require.ErrorIs(t, err, apierrors.ErrBadRequest)
}

func TestUpdateStudentProfileTooYoung(t *testing.T) {
	env := newTestService(t)

	birth := time.Now().AddDate(-15, 0, 0)
	profile := models.Profile{
		BirthDate: &birth,
	}
	err := env.svc.UpdateStudentProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"), profile)
	require.ErrorIs(t, err, apierrors.ErrUserMustBe16)
}

func TestUpdateStudentProfileNoClaims(t *testing.T) {
	env := newTestService(t)

	fname := "Ivan"
	err := env.svc.UpdateStudentProfile(context.Background(), models.Profile{FirstName: &fname})
	require.ErrorIs(t, err, apierrors.ErrInternalServer)
}

func TestUpdateStudentProfileSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	fname := "Ivan"
	lname := "Ivanov"
	cityID := 77

	env.mock.ExpectExec(`UPDATE interns SET first_name = $2 WHERE user_id = $1`).
		WithArgs(1, fname).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := env.svc.UpdateStudentProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"), models.Profile{
		FirstName: &fname,
	})
	require.NoError(t, err)

	env.mock.ExpectExec(`UPDATE interns SET first_name = $2, last_name = $3, city_id = $4 WHERE user_id = $1`).
		WithArgs(1, fname, lname, cityID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = env.svc.UpdateStudentProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"), models.Profile{
		FirstName: &fname,
		LastName:  &lname,
		CityID:    &cityID,
	})
	require.NoError(t, err)
}

func TestGetMyProfileSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT id FROM interns WHERE user_id = $1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	env.mock.ExpectQuery(`SELECT i.id, i.user_id, i.first_name, i.last_name, i.birth_date, i.city_id, i.university, i.degree, i.bio, i.experience, i.image FROM interns i WHERE i.id = $1 AND NOT EXISTS(SELECT 1 FROM user_bans ub WHERE ub.user_id = i.user_id)`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "first_name", "last_name", "birth_date", "city_id", "university", "degree", "bio", "experience", "image"}).
			AddRow(5, 1, "Ivan", "Ivanov", time.Now().AddDate(-20, 0, 0), 77, "MSTU", "BSc", "bio", "1 year", "avatar.png"))

	env.mock.ExpectQuery(`SELECT id FROM interns WHERE user_id = $1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	env.mock.ExpectQuery(`SELECT skill_id FROM intern_skills WHERE intern_id = $1`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"skill_id"}))

	resp, err := env.svc.GetMyProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"))
	require.NoError(t, err)
	require.Equal(t, "intern@test.ru", resp.Email)
	require.Equal(t, "Ivan", *resp.Profile.FirstName)
	require.Equal(t, 77, *resp.Profile.CityID)
	require.Empty(t, resp.Skills)
}

func TestGetMyProfileProfileNotFound(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectQuery(`SELECT id FROM interns WHERE user_id = $1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	env.mock.ExpectQuery(`SELECT i.id, i.user_id, i.first_name, i.last_name, i.birth_date, i.city_id, i.university, i.degree, i.bio, i.experience, i.image FROM interns i WHERE i.id = $1 AND NOT EXISTS(SELECT 1 FROM user_bans ub WHERE ub.user_id = i.user_id)`).
		WithArgs(5).
		WillReturnError(sql.ErrNoRows)

	_, err := env.svc.GetMyProfile(ctxWithClaims(1, constants.Intern, "intern@test.ru"))
	require.ErrorIs(t, err, apierrors.ErrProfileNotFound)
}
