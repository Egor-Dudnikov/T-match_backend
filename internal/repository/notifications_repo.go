package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (r *Repository) GetMyNotifications(ctx context.Context, id int) ([]models.Notification, error) {
	notifications := []models.Notification{}

	query := `
        SELECT 
            n.id,
            n.user_id,
            n.type,
            n.is_read,
            n.created_at,
            sc.id as sc_id,
            sc.notification_id as sc_notification_id,
            sc.internship_id as sc_internship_id,
            sc.company_id as sc_company_id,
            sc.new_status,
            inv.id as inv_id,
            inv.notification_id as inv_notification_id,
            inv.internship_id as inv_internship_id,
            inv.company_id as inv_company_id,
            inv.message
        FROM notifications n
        LEFT JOIN change_status_data sc ON n.id = sc.notification_id
        LEFT JOIN invate_data inv ON n.id = inv.notification_id
        WHERE n.user_id = $1
        ORDER BY n.created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return notifications, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var n models.Notification
		var scID, scNotificationID, scInternshipID, scCompanyID sql.NullInt64
		var invID, invNotificationID, invInternshipID, invCompanyID sql.NullInt64
		var newStatus, message sql.NullString

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.IsRead,
			&n.CreatedAt,
			&scID,
			&scNotificationID,
			&scInternshipID,
			&scCompanyID,
			&newStatus,
			&invID,
			&invNotificationID,
			&invInternshipID,
			&invCompanyID,
			&message,
		)
		if err != nil {
			return nil, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}

		if scID.Valid {
			n.Data = models.ChangeStatusData{
				ID:             int(scID.Int64),
				NotificationID: int(scNotificationID.Int64),
				InternshipID:   int(scInternshipID.Int64),
				CompanyID:      int(scCompanyID.Int64),
				NewStatus:      newStatus.String,
			}
		} else if invID.Valid {
			var msg *string
			if message.Valid {
				msg = &message.String
			}
			n.Data = models.InvateData{
				ID:             int(invID.Int64),
				NotificationID: int(invNotificationID.Int64),
				InternshipID:   int(invInternshipID.Int64),
				CompanyID:      int(invCompanyID.Int64),
				Message:        msg,
			}
		} else {
			return notifications, apierrors.ErrInternalServer
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *Repository) SetReadStatusOfNotification(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = TRUE WHERE user_id = $1`, userID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) NewChangeStatusNotification(ctx context.Context, responseID int, internshipID int, newStatus string) (models.Notification, error) {
	var notification models.Notification
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			if err != nil {
				err = fmt.Errorf("original error: %w, rollback error: %w", err, rbErr)
			} else {
				err = apierrors.Wrap(apierrors.ErrDatabaseError, rbErr)
			}
		}
	}()

	var notificationID int
	var createdAt time.Time
	var userID int

	queryNotif := `
        INSERT INTO notifications (user_id, type, created_at)
        VALUES ((SELECT user_id FROM interns WHERE id = (SELECT intern_id FROM applications WHERE id = $1)), $2, NOW())
        RETURNING id, user_id, created_at
    `
	err = tx.QueryRowContext(
		ctx,
		queryNotif,
		responseID,
		constants.ChangeStatusType,
	).Scan(&notificationID, &userID, &createdAt)
	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	queryData := `
        INSERT INTO change_status_data (
            notification_id,
            internship_id,
            company_id,
            new_status
        ) VALUES ($1, $2, (SELECT company_id FROM internships WHERE id = $2), $3)
		RETURNING id, company_id
    `

	var dataID int
	var companyID int

	err = tx.QueryRowContext(
		ctx,
		queryData,
		notificationID,
		internshipID,
		newStatus,
	).Scan(&dataID, &companyID)

	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	if err := tx.Commit(); err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	notification = models.Notification{
		ID:        notificationID,
		UserID:    userID,
		Type:      constants.ChangeStatusType,
		IsRead:    false,
		CreatedAt: createdAt,
		Data: models.ChangeStatusData{
			ID:             dataID,
			NotificationID: notificationID,
			InternshipID:   internshipID,
			CompanyID:      companyID,
			NewStatus:      newStatus,
		},
	}

	return notification, nil
}

func (r *Repository) NewInviteNotification(ctx context.Context, userID int, internshipID int, companyID int, message *string) (models.Notification, error) {
	var notification models.Notification

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			if err != nil {
				err = fmt.Errorf("original error: %w, rollback error: %w", err, rbErr)
			} else {
				err = apierrors.Wrap(apierrors.ErrDatabaseError, rbErr)
			}
		}
	}()

	var notificationID int
	var createdAt time.Time

	queryNotif := `
        INSERT INTO notifications (user_id, type, created_at)
        VALUES ($1, $2, NOW())
		RETURNING id, created_at
    `

	err = tx.QueryRowContext(
		ctx,
		queryNotif,
		userID,
		constants.InvateType,
	).Scan(&notificationID, &createdAt)

	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	var dataID int

	queryData := `
        INSERT INTO invate_data (
            notification_id,
            internship_id,
            company_id,
            message
        ) VALUES ($1, $2, $3, $4)
		 RETURNING id
    `
	err = tx.QueryRowContext(
		ctx,
		queryData,
		notificationID,
		internshipID,
		companyID,
		message,
	).Scan(&dataID)

	if err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	if err := tx.Commit(); err != nil {
		return notification, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	notification = models.Notification{
		ID:        notificationID,
		UserID:    userID,
		Type:      constants.InvateType,
		IsRead:    false,
		CreatedAt: createdAt,
		Data: models.InvateData{
			ID:             dataID,
			NotificationID: notificationID,
			InternshipID:   internshipID,
			CompanyID:      companyID,
			Message:        message,
		},
	}

	return notification, nil
}
