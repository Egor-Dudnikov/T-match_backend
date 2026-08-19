// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"log"

	"github.com/lib/pq"
)

// GetMyResponses returns the applications of the intern with the given ID, excluding banned users.
func (r *Repository) GetMyResponses(ctx context.Context, internID int) ([]models.Response, error) {
	res := []models.Response{}

	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.intern_id, a.internship_id, a.status, a.created_at
		FROM applications a
		JOIN interns i ON a.intern_id = i.id
		WHERE a.intern_id = $1
		  AND NOT EXISTS(
			  SELECT 1 FROM user_bans ub 
			  WHERE ub.user_id = i.user_id
		  )
	`, internID)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("repo: close rows: %v", cerr)
		}
	}()

	for rows.Next() {
		response := models.Response{}
		err = rows.Scan(
			&response.ID,
			&response.InternID,
			&response.InternshipID,
			&response.Status,
			&response.CreatedAt,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		res = append(res, response)
	}
	return res, nil
}

// RespondInternship creates an application from the intern to the internship and returns its ID.
func (r *Repository) RespondInternship(ctx context.Context, internID int, internshipID int) (int, error) {
	var respondID int
	err := r.db.QueryRowContext(ctx, `INSERT INTO applications (intern_id, internship_id, created_at) VALUES ($1, $2, NOW()) RETURNING id`, internID, internshipID).Scan(&respondID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return respondID, apierrors.ErrAlreadyResponded
		}
		return respondID, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return respondID, nil
}

// InternshipsResponse returns the applications for the internship with the given ID and sets them to reviewing.
func (r *Repository) InternshipsResponse(ctx context.Context, internshipID int) ([]models.Response, error) {
	res := []models.Response{}
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM applications WHERE internship_id = $1`, internshipID)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("repo: close rows: %v", cerr)
		}
	}()
	for rows.Next() {
		response := models.Response{}
		err := rows.Scan(&response.ID,
			&response.InternID,
			&response.InternshipID,
			&response.Status,
			&response.CreatedAt)

		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		res = append(res, response)
	}
	err = r.SetReviewStatus(ctx, internshipID)
	return res, err
}

// DeleteRespondInternship deletes the pending application of the intern to the internship.
func (r *Repository) DeleteRespondInternship(ctx context.Context, internID int, internshipID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM applications WHERE internship_id = $1 AND intern_id = $2 AND status = 'pending'", internshipID, internID)
	return err
}

// SetReviewStatus sets the status to reviewing for all pending applications of the internship.
func (r *Repository) SetReviewStatus(ctx context.Context, internshipID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE applications SET status = 'reviewing' WHERE status = 'pending' AND internship_id = $1`, internshipID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

// SetResponseStatus sets the status of the application with the given ID.
func (r *Repository) SetResponseStatus(ctx context.Context, ID int, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE applications SET status = $1 WHERE id = $2`, status, ID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

// GetInternshipIDByResponseID returns the internship ID of the application with the given ID.
func (r *Repository) GetInternshipIDByResponseID(ctx context.Context, ResponseID int) (int, error) {
	var internshipID int
	err := r.db.QueryRowContext(ctx, `SELECT internship_id FROM applications WHERE id = $1`, ResponseID).Scan(&internshipID)
	if err != nil {
		return internshipID, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return internshipID, nil
}
