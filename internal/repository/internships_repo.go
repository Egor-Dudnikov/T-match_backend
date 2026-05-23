// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (r *Repository) NewInternship(ctx context.Context, interships models.Internship, userID int) (int, error) {
	id, err := r.GetCmpanyIdByUserId(ctx, userID)
	if err != nil {
		return 0, err
	}

	var internshipID int
	err = r.db.QueryRowContext(ctx, `INSERT INTO internships (company_id, title, description, salary, duration_months, location, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id;`, id, interships.Title, interships.Description, interships.Salary, interships.DurationMonth, interships.Location).Scan(&internshipID)
	return internshipID, err
}

func (r *Repository) GetInternshipById(ctx context.Context, id int) (models.Internship, error) {
	internship := models.Internship{}
	err := r.db.QueryRowContext(ctx, `SELECT id, company_id, title, description, 
           salary, duration_months, location, 
           created_at, is_archived FROM internships FROM internships WHERE id = $1 AND is_archived = FALSE`, id).Scan(
		&internship.Id,
		&internship.CompanyId,
		&internship.Title,
		&internship.Description,
		&internship.Salary,
		&internship.DurationMonth,
		&internship.Location,
		&internship.CreatedAt,
		&internship.IsArchived,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internship, apierrors.ErrInternshipNotFound
		}
		return internship, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

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

	query.WriteString(" WHERE id = $1 AND is_archived = FALSE")

	_, err := r.db.ExecContext(ctx, query.String(), values...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apierrors.ErrInternshipNotFound
		}
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	return err
}

func (r *Repository) GetCompanyIdByInternshipId(ctx context.Context, id int) (int, error) {
	var companyId int
	err := r.db.QueryRowContext(ctx, `SELECT company_id FROM internships WHERE id = $1`, id).Scan(&companyId)
	return companyId, err
}

func (r *Repository) ArchivedInternship(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE internships SET is_archived = TRUE`)
	return err
}

func (r *Repository) SearchInternship(ctx context.Context, filters models.SearchInternship) ([]models.Internship, error) {
	res := []models.Internship{}
	query := `SELECT id, company_id, title, description, 
           salary, duration_months, location, 
           created_at, is_archived FROM internships
				WHERE (tsv_content @@ plainto_tsquery('russian', $1)) AND is_archived = FALSE;
	`

	rows, err := r.db.QueryContext(ctx, query, filters.Query)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		internship := models.Internship{}
		rows.Scan(
			&internship.Id,
			&internship.CompanyId,
			&internship.Title,
			&internship.Description,
			&internship.Salary,
			&internship.DurationMonth,
			&internship.Location,
			&internship.CreatedAt,
			&internship.IsArchived,
		)
		res = append(res, internship)
	}

	return res, nil
}
