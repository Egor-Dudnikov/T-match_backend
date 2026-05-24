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
	var query strings.Builder
	query.WriteString("SELECT id, company_id, title, description, salary, duration_months, location, created_at, is_archived FROM internships")

	index := 1
	values := []interface{}{}
	query.WriteString(" WHERE is_archived = FALSE")
	if filters.Query != nil {
		query.WriteString(" AND ")
		query.WriteString("tsv_content @@ to_tsquery('russian', $")
		query.WriteString(strconv.Itoa(index))
		query.WriteString(")")
		values = append(values, *filters.Query)
		index++
	}

	if filters.SalaryMax != nil {
		query.WriteString(" AND ")
		query.WriteString("salary <= $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.SalaryMax)
		index++
	}

	if filters.SalaryMin != nil {
		query.WriteString(" AND ")
		query.WriteString("salary >= $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.SalaryMin)
		index++
	}

	if filters.DurationMin != nil {
		query.WriteString(" AND ")
		query.WriteString("duration_months >= $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.DurationMin)
		index++
	}

	if filters.DurationMax != nil {
		query.WriteString(" AND ")
		query.WriteString("duration_months <= $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.DurationMax)
		index++
	}

	if filters.Location != nil {
		query.WriteString(" AND ")
		query.WriteString("location ILIKE $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, (fmt.Sprintf("%c%s%c", '%', *filters.Location, '%')))
		index++
	}

	query.WriteString(" ORDER BY ")
	delimiter := false
	if filters.Offset != nil && filters.Sort != nil && sortValid(*filters.Sort) {
		query.WriteString(*filters.Sort)

		if *filters.Order == 1 {
			query.WriteString(" ASC")
		} else {
			query.WriteString(" DESC")
		}
		delimiter = true
	}

	if filters.Query != nil {
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("ts_rank(tsv_content, to_tsquery('russian', $1)) DESC")
		delimiter = true
	}

	if delimiter {
		query.WriteString(", ")
	}
	query.WriteString("created_at DESC")

	if filters.Limit != nil {
		query.WriteString(" LIMIT $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.Limit)
		index++
	}

	if filters.Offset != nil {
		query.WriteString(" OFFSET $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, *filters.Offset)
		index++
	}

	query.WriteString(";")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

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

func sortValid(sort string) bool {
	if sort == "salary" {
		return true
	} else if sort == "duration_month" {
		return true
	}
	return false
}
