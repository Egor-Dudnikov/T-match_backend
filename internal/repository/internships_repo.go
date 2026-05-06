package repository

import (
	"T-match_backend/internal/models"
	"context"
	"strconv"
	"strings"
)

func (r *Repository) NewInternship(ctx context.Context, interships models.Internship, userID int) error {
	id, err := r.GetCmpanyIdByUserId(ctx, userID)
	if err != nil {
		return err
	}
	err = r.db.QueryRowContext(ctx, `INSERT INTO internships (company_id, title, description, salary, duration_months, location, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW())`, id, interships.Title, interships.Description, interships.Salary, interships.DurationMonth, interships.Location).Err()
	return err
}

func (r *Repository) GetInternshipById(ctx context.Context, id int) (models.Internship, error) {
	internship := models.Internship{}
	err := r.db.QueryRowContext(ctx, `SELECT * FROM internships WHERE id = $1`, id).Scan(
		&internship.Id,
		&internship.CompanyId,
		&internship.Title,
		&internship.Description,
		&internship.Salary,
		&internship.Location,
		&internship.IsArchived,
		&internship.DurationMonth,
		&internship.CreatedAt,
	)
	return internship, err
}

func (r *Repository) UpdateInternships(ctx context.Context, internship models.InternshipUpdate) error {
	var query strings.Builder
	cnt := 1
	delimiter := false
	values := []interface{}{internship.Id}
	query.WriteString("UPDATE internships SET ")

	addFilled := func(filled string, value any) {
		cnt++
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString(filled)
		query.WriteString(" = $")
		query.WriteString(strconv.Itoa(cnt))
		delimiter = true
		values = append(values, value)
	}

	if internship.Title != "" {
		addFilled("title", internship.Title)
	}
	if internship.Description != "" {
		addFilled("description", internship.Description)
	}
	if internship.Location != "" {
		addFilled("location", internship.Location)
	}
	if internship.Salary != nil {
		addFilled("salary", *internship.Salary)
	}
	if internship.DurationMonth != 0 {
		addFilled("duration_month", internship.DurationMonth)
	}

	query.WriteString(" WHERE id = $1")

	_, err := r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetCompanyIdByInternshipId(ctx context.Context, id int) (int, error) {
	var companyId int
	err := r.db.QueryRowContext(ctx, `SELECT company_id FROM internships WHERE id = $1`, id).Scan(&companyId)
	return companyId, err
}

func (r *Repository) ArchivedInternship(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE internships SET is_archived = FALSE`)
	return err
}
