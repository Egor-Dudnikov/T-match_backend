// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetMyNotificationsNoClaims(t *testing.T) {
	env := newTestService(t)

	_, err := env.svc.GetMyNotifications(context.Background())
	require.ErrorIs(t, err, apierrors.ErrInternalServer)
}

func TestGetMyNotificationsSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	created := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	env.mock.ExpectQuery(`SELECT n.id, n.user_id, n.type, n.is_read, n.created_at, sc.id as sc_id, sc.notification_id as sc_notification_id, sc.internship_id as sc_internship_id, sc.company_id as sc_company_id, sc.new_status, inv.id as inv_id, inv.notification_id as inv_notification_id, inv.internship_id as inv_internship_id, inv.company_id as inv_company_id, inv.message, na.id as na_id, na.notification_id as na_notification_id, na.internship_id as na_internship_id, na.intern_id as na_intern_id, na.response_id as na_response_id FROM notifications n LEFT JOIN change_status_data sc ON n.id = sc.notification_id LEFT JOIN invate_data inv ON n.id = inv.notification_id LEFT JOIN new_application_data na ON n.id = na.notification_id WHERE n.user_id = $1 ORDER BY n.created_at DESC`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "type", "is_read", "created_at",
			"sc_id", "sc_notification_id", "sc_internship_id", "sc_company_id", "sc_new_status",
			"inv_id", "inv_notification_id", "inv_internship_id", "inv_company_id", "inv_message",
			"na_id", "na_notification_id", "na_internship_id", "na_intern_id", "na_response_id",
		}).
			AddRow(10, 5, constants.InvateType, false, created,
				nil, nil, nil, nil, nil,
				100, 10, 55, 3, "hello",
				nil, nil, nil, nil, nil))

	notifications, err := env.svc.GetMyNotifications(ctxWithClaims(5, constants.Intern, "intern@test.ru"))
	require.NoError(t, err)
	require.Len(t, notifications, 1)

	n := notifications[0]
	require.Equal(t, 10, n.ID)
	require.Equal(t, 5, n.UserID)
	require.Equal(t, constants.InvateType, n.Type)
	require.Equal(t, created, n.CreatedAt)

	invite, ok := n.Data.(models.InvateData)
	require.True(t, ok)
	require.Equal(t, 100, invite.ID)
	require.Equal(t, 55, invite.InternshipID)
	require.Equal(t, 3, invite.CompanyID)
	require.Equal(t, "hello", *invite.Message)
}

func TestSetReadStatusOfNotificationNoClaims(t *testing.T) {
	env := newTestService(t)

	err := env.svc.SetReadStatusOfNotification(context.Background())
	require.ErrorIs(t, err, apierrors.ErrInternalServer)
}

func TestSetReadStatusOfNotificationSuccess(t *testing.T) {
	env := newTestService(t)
	defer func() {
		require.NoError(t, env.mock.ExpectationsWereMet(), "unmatched sql expectations")
	}()

	env.mock.ExpectExec(`UPDATE notifications SET is_read = TRUE WHERE user_id = $1`).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := env.svc.SetReadStatusOfNotification(ctxWithClaims(5, constants.Intern, "intern@test.ru"))
	require.NoError(t, err)
}
