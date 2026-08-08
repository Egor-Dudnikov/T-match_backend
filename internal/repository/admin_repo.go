package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
)

func (r *Repository) GetStats(ctx context.Context) (models.AdminStats, error) {
	var stats models.AdminStats

	err := r.db.QueryRowContext(ctx, `
        SELECT 
            (SELECT COUNT(*) FROM interns) AS total_interns,
            (SELECT COUNT(*) FROM companies) AS total_companies,
            (SELECT COUNT(*) FROM internships WHERE is_archived = false) AS total_internships,
            (SELECT COUNT(*) FROM applications) AS total_responses
    `).Scan(
		&stats.TotalInterns,
		&stats.TotalCompanies,
		&stats.TotalInternships,
		&stats.TotalResponses,
	)
	if err != nil {
		return stats, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	err = r.db.QueryRowContext(ctx, `
        SELECT 
            COUNT(*) FILTER (WHERE status = 'pending') AS pending,
            COUNT(*) FILTER (WHERE status = 'reviewing') AS reviewing,
            COUNT(*) FILTER (WHERE status = 'accepted') AS accepted,
            COUNT(*) FILTER (WHERE status = 'rejected') AS rejected
        FROM applications
    `).Scan(
		&stats.ResponsesPending,
		&stats.ResponsesReviewing,
		&stats.ResponsesAccepted,
		&stats.ResponsesRejected,
	)
	if err != nil {
		return stats, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	err = r.db.QueryRowContext(ctx, `
        SELECT 
            (SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '7 days') AS new_users,
            (SELECT COUNT(*) FROM internships WHERE created_at > NOW() - INTERVAL '7 days') AS new_internships,
            (SELECT COUNT(*) FROM applications WHERE created_at > NOW() - INTERVAL '7 days') AS new_responses
    `).Scan(
		&stats.NewUsers7Days,
		&stats.NewInternships7Days,
		&stats.NewResponses7Days,
	)
	if err != nil {
		return stats, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	return stats, nil
}

func (r *Repository) BanUser(ctx context.Context, userID int, bannedBy int, reason string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO user_bans (user_id, reason, banned_by)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE
        SET reason = $2, banned_by = $3, banned_at = NOW()
    `, userID, reason, bannedBy)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) UnbanUser(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM user_bans WHERE user_id = $1
    `, userID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) IsUserBanned(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id = $1)
    `, userID).Scan(&exists)
	if err != nil {
		return false, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return exists, nil
}

func (r *Repository) GetUserBanReason(ctx context.Context, userID int) (string, error) {
	var reason string
	err := r.db.QueryRowContext(ctx, `
        SELECT reason FROM user_bans WHERE user_id = $1
    `, userID).Scan(&reason)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return reason, nil
}

func (r *Repository) InternshipExists(ctx context.Context, internshipID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM internships WHERE id = $1)`, internshipID).Scan(&exists)
	if err != nil {
		return false, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return exists, nil
}

func (r *Repository) AdminDeleteInternship(ctx context.Context, internshipID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM internships WHERE id = $1`, internshipID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}
