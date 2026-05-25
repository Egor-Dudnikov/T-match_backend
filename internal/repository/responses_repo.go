// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"

	"github.com/lib/pq"
)

func (r *Repository) GetMyResponses(ctx context.Context, internID int) ([]models.Response, error) {
	res := []models.Response{}
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM applications WHERE intern_id = $1`, internID)
	if err != nil {
		return res, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()
	for rows.Next() {
		response := models.Response{}
		rows.Scan(&response.ID,
			&response.InternID,
			&response.InternshipID,
			&response.Status,
			&response.CreatedAt)
		res = append(res, response)
	}
	return res, nil
}

func (r *Repository) RespondInternship(ctx context.Context, internID int, internshipID int) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO applications (intern_id, internship_id, created_at) VALUES ($1, $2, NOW())`, internID, internshipID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return apierrors.ErrAlreadyResponded
		}
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) InternshipsResponse(ctx context.Context, internshipID int) ([]models.Response, error) {
	res := []models.Response{}
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM applications WHERE internship_id = $1`, internshipID)
	if err != nil {
		return res, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()
	for rows.Next() {
		response := models.Response{}
		rows.Scan(&response.ID,
			&response.InternID,
			&response.InternshipID,
			&response.Status,
			&response.CreatedAt)
		res = append(res, response)
	}
	err = r.SetReviewStatus(ctx, internshipID)
	return res, err
}

func (r *Repository) SetReviewStatus(ctx context.Context, internshipID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE applications SET status = 'reviewing' WHERE status = 'pending' AND internship_id = $1`, internshipID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) SetResponseStatus(ctx context.Context, ID int, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE applications SET status = $1 WHERE id = $2`, status, ID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) GetInternshipIdByResponseId(ctx context.Context, ResponseID int) (int, error) {
	var internshipID int
	err := r.db.QueryRowContext(ctx, `SELECT internship_id FROM applications WHERE id = $1`, ResponseID).Scan(&internshipID)
	if err != nil {
		return internshipID, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return internshipID, nil
}
