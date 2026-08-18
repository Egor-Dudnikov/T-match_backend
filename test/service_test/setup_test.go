// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"T-match_backend/internal/cache"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/dadata"
	"T-match_backend/internal/models"
	"T-match_backend/internal/recsys"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/s3"
	"T-match_backend/internal/service"
	"T-match_backend/internal/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	svc   *service.Service
	mock  sqlmock.Sqlmock
	redis *miniredis.Miniredis
}

func newTestService(t *testing.T) *testEnv {
	t.Helper()

	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if normalizeSQL(expectedSQL) == normalizeSQL(actualSQL) {
			return nil
		}
		return fmt.Errorf("sql mismatch: expected %q, got %q", expectedSQL, actualSQL)
	})

	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	require.NoError(t, err)

	repo := repository.NewRepository(mockDB)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	myCache := cache.NewRedis(redisClient)

	validate, err := utils.NewValidator()
	require.NoError(t, err)

	email := service.NewEmailClient(models.EmailConfig{})
	storage := &s3.Storage{}
	daData := &dadata.Client{}
	hub := &service.Hub{}
	recsysClient := recsys.NewClient("")

	app := service.Newservice(repo, myCache, email, validate, storage, daData, hub, recsysClient)

	return &testEnv{
		svc:   app,
		mock:  mock,
		redis: mr,
	}
}

// ctxWithClaims returns a context carrying the given claims payload.
func ctxWithClaims(userID int, role, email string) context.Context {
	claims := models.Claims{
		UserID:   userID,
		Role:     role,
		Email:    email,
		DeviceID: "device-12345",
	}
	return context.WithValue(context.Background(), constants.ClaimsKey, claims)
}

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-tests")
	os.Exit(m.Run())
}

// normalizeSQL removes all whitespace so that multiline repository queries
// (including `NOT EXISTS( SELECT ... )` spacing) match single-line expectations.
func normalizeSQL(q string) string {
	return strings.Join(strings.Fields(q), "")
}
