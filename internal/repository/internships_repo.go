// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/lib/pq"
)

// NewInternship creates a new internship for the company of the given user and returns its ID.
func (r *Repository) NewInternship(ctx context.Context, interships models.Internship, userID int) (int, error) {
	id, err := r.GetCompanyIDByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	var internshipID int
	err = r.db.QueryRowContext(ctx, `INSERT INTO internships (company_id, title, description, salary, duration_months, city_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id;`, id, interships.Title, interships.Description, interships.Salary, interships.DurationMonth, interships.CityID).Scan(&internshipID)
	if err != nil {
		return 0, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return internshipID, nil
}

// GetInternshipByID returns the non-archived internship with the given ID, excluding banned companies.
func (r *Repository) GetInternshipByID(ctx context.Context, id int) (models.Internship, error) {
	internship := models.Internship{}
	err := r.db.QueryRowContext(ctx, `
		SELECT i.id, i.company_id, i.title, i.description, 
		       i.salary, i.duration_months, i.city_id, 
		       i.created_at, i.is_archived 
		FROM internships i
		JOIN companies comp ON i.company_id = comp.id
		WHERE i.id = $1 
		  AND i.is_archived = FALSE
		  AND NOT EXISTS(
			  SELECT 1 FROM user_bans ub 
			  WHERE ub.user_id = comp.user_id
		  )
	`, id).Scan(
		&internship.ID,
		&internship.CompanyID,
		&internship.Title,
		&internship.Description,
		&internship.Salary,
		&internship.DurationMonth,
		&internship.CityID,
		&internship.CreatedAt,
		&internship.IsArchived,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internship, apierrors.ErrInternshipNotFound
		}
		return internship, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	return internship, nil
}

// UpdateInternships updates the filled fields of the non-archived internship with the given ID.
func (r *Repository) UpdateInternships(ctx context.Context, internship models.InternshipUpdate) error {

	query := newUpdateQuery("UPDATE internships SET ", internship.ID)

	addFilled[string](query, "title", internship.Title)
	addFilled[string](query, "description", internship.Description)
	addFilled[int](query, "city_id", internship.CityID)
	addFilled[int](query, "salary", internship.Salary)
	addFilled[int](query, "duration_months", internship.DurationMonth)

	if query.empty() {
		return nil
	}

	queryStr, values := query.parseBuilder(" WHERE id = $1 AND is_archived = FALSE")

	_, err := r.db.ExecContext(ctx, queryStr, values...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apierrors.ErrInternshipNotFound
		}
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}

	return err
}

// GetCompanyIDByInternshipID returns the company ID of the non-archived internship with the given ID.
func (r *Repository) GetCompanyIDByInternshipID(ctx context.Context, id int) (int, error) {
	var companyID int
	err := r.db.QueryRowContext(ctx, `SELECT company_id FROM internships WHERE id = $1 AND is_archived = FALSE`, id).Scan(&companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return companyID, apierrors.Wrap(apierrors.ErrInternshipNotFound, err)
		}
		return companyID, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return companyID, nil
}

// GetCompanyInternships returns the internships of the company with the given ID, optionally only non-archived ones.
func (r *Repository) GetCompanyInternships(ctx context.Context, id int, hintArchiveInternships bool) ([]models.Internship, error) {
	res := []models.Internship{}
	query := `SELECT i.id, i.company_id, i.title, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived FROM internships i WHERE i.company_id = $1`
	if hintArchiveInternships {
		query += " AND i.is_archived = FALSE"
	}
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("repo: close rows: %v", cerr)
		}
	}()

	for rows.Next() {
		internship := models.Internship{}
		err = rows.Scan(
			&internship.ID,
			&internship.CompanyID,
			&internship.Title,
			&internship.Salary,
			&internship.DurationMonth,
			&internship.CityID,
			&internship.CreatedAt,
			&internship.IsArchived,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		res = append(res, internship)
	}

	return res, nil
}

// ArchivedInternship marks the internship with the given ID as archived.
func (r *Repository) ArchivedInternship(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE internships SET is_archived = TRUE WHERE id = $1`, id)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

// GetInternshipsByIDs returns the non-archived internships with the given IDs, excluding banned companies,
// ordered by the order of the IDs.
func (r *Repository) GetInternshipsByIDs(ctx context.Context, ids []int) ([]models.Internship, error) {
	res := []models.Internship{}
	if len(ids) == 0 {
		return res, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT i.id, i.company_id, i.title, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived
		FROM internships i
		WHERE i.id = ANY($1)
		  AND i.is_archived = FALSE
		  AND NOT EXISTS(
			  SELECT 1 FROM user_bans ub
			  JOIN companies comp ON ub.user_id = comp.user_id
			  WHERE comp.id = i.company_id
		  )
	`, pq.Array(ids))
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("repo: close rows: %v", cerr)
		}
	}()

	byID := make(map[int]models.Internship, len(ids))
	for rows.Next() {
		internship := models.Internship{}
		err = rows.Scan(
			&internship.ID,
			&internship.CompanyID,
			&internship.Title,
			&internship.Salary,
			&internship.DurationMonth,
			&internship.CityID,
			&internship.CreatedAt,
			&internship.IsArchived,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		byID[internship.ID] = internship
	}

	for _, id := range ids {
		if internship, ok := byID[id]; ok {
			res = append(res, internship)
		}
	}
	return res, nil
}

// SearchInternship returns non-archived internships matching the given filters.
func (r *Repository) SearchInternship(ctx context.Context, filters models.SearchInternship) ([]models.Internship, error) {
	res := []models.Internship{}
	query := newQuerySelectBuilder("SELECT i.id, i.company_id, i.title, i.salary, i.duration_months, i.city_id, i.created_at, i.is_archived FROM internships i")

	query.addWhere(`
		i.is_archived = FALSE 
		AND NOT EXISTS(
			SELECT 1 FROM user_bans ub 
			JOIN companies comp ON ub.user_id = comp.user_id 
			WHERE comp.id = i.company_id
		)
	`)

	addWhereWithIndex[string](query, "tsv_content @@ to_tsquery('russian', $", ")", filters.Query)
	addWhereWithIndex[int](query, "salary <= $", "", filters.SalaryMax)
	addWhereWithIndex[int](query, "salary >= $", "", filters.SalaryMin)
	addWhereWithIndex[int](query, "duration_months >= $", "", filters.DurationMin)
	addWhereWithIndex[int](query, "duration_months <= $", "", filters.DurationMax)

	if filters.Skills != nil {
		addWhereWithIndexes(query, "(SELECT COUNT(DISTINCT skill_id) FROM internship_skills WHERE internship_id = i.id AND skill_id ",
			") = "+strconv.Itoa(len(*filters.Skills)), filters.Skills)
	}

	addWhereWithIndex[int](query, "i.city_id = $", "", filters.CityID)

	if sortValid(filters.Order, filters.Sort) {
		if *filters.Order == 1 {
			query.addOrderBy(*filters.Sort, constants.ASC)
		} else {
			query.addOrderBy(*filters.Sort, constants.DESC)
		}
	}

	if filters.Query != nil {
		query.addOrderBy("ts_rank(tsv_content, to_tsquery('russian', $1)) ", constants.DESC)
	}

	query.addOrderBy("created_at ", constants.DESC)

	query.addLimit(filters.Limit)
	query.addOffset(filters.Offset)

	queryStr, values := query.parseBuilder()

	rows, err := r.db.QueryContext(ctx, queryStr, values...)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("repo: close rows: %v", cerr)
		}
	}()

	for rows.Next() {
		internship := models.Internship{}
		err = rows.Scan(
			&internship.ID,
			&internship.CompanyID,
			&internship.Title,
			&internship.Salary,
			&internship.DurationMonth,
			&internship.CityID,
			&internship.CreatedAt,
			&internship.IsArchived,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		res = append(res, internship)
	}

	return res, nil
}

func sortValid(order *int, sort *string) bool {
	if order == nil || sort == nil {
		return false
	}
	switch *sort {
	case "salary", "duration_months":
		return true
	default:
		return false
	}
}
