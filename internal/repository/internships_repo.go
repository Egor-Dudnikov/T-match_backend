package repository

import (
	"T-match_backend/internal/models"
	"context"
)

func (r *Repository) NewInternship(ctx context.Context, interships models.Internship, userID int) error {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM companies WHERE user_id = $1`, userID).Scan(&id)
	if err != nil {
		return err
	}
	err = r.db.QueryRowContext(ctx, `INSERT INTO internships (company_id, title, description, salary, duration_months, location, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW())`, id, interships.Title, interships.Description, interships.Salary, interships.DurationMonth, interships.Location).Err()
	return err
}
