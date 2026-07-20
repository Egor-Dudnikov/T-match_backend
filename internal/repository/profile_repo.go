// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (r *Repository) QueryProfile(ctx context.Context, id int, profile models.Profile) error {
	var query strings.Builder
	query.WriteString("UPDATE interns SET ")
	delimiter := false
	cnt := 1
	values := []interface{}{id}

	addFilled := func(name string, value any) {
		if delimiter {
			query.WriteString(", ")
		}
		cnt++

		query.WriteString(name)
		query.WriteString(" = $")
		query.WriteString(strconv.Itoa(cnt))

		values = append(values, value)
		delimiter = true
	}

	if profile.FirstName != nil {
		addFilled("first_name", &profile.FirstName)
	}

	if profile.LastName != nil {
		addFilled("last_name", &profile.LastName)
	}

	if profile.BirthDate != nil {
		addFilled("birth_date", &profile.BirthDate)
	}

	if profile.Bio != nil {
		addFilled("bio", &profile.Bio)
	}

	if profile.Degree != nil {
		addFilled("degree", &profile.Degree)
	}

	if profile.Experience != nil {
		addFilled("experience", &profile.Experience)
	}

	if profile.Location != nil {
		addFilled("location", &profile.Location)
	}

	if profile.University != nil {
		addFilled("university", &profile.University)
	}

	query.WriteString(" WHERE user_id = $1")

	if cnt == 1 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, query.String(), values...)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) GetProfileIDByUserID(ctx context.Context, userID int) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM interns WHERE user_id = $1`, userID).Scan(&id)
	if err != nil {
		return id, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return id, err
}

func (r *Repository) GetProfile(ctx context.Context, id int) (models.Profile, error) {
	profile := models.Profile{}
	err := r.db.QueryRowContext(ctx, `SELECT id, first_name, last_name, birth_date, location, university, degree, bio, experience, image 
	FROM interns
	WHERE id = $1`, id).Scan(
		&profile.ID,
		&profile.FirstName,
		&profile.LastName,
		&profile.BirthDate,
		&profile.Location,
		&profile.University,
		&profile.Degree,
		&profile.Bio,
		&profile.Experience,
		&profile.Image)
	if err != nil {
		return profile, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return profile, nil
}

func (r *Repository) SetMyAvatar(ctx context.Context, url string, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE interns SET
	image = $2
	WHERE user_id = $1`, id, url)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) GetAllSkills(ctx context.Context) ([]models.Skill, error) {
	skills := []models.Skill{}
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM skills;`)
	if err != nil {
		return skills, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		skills = append(skills, models.Skill{ID: id, Name: name})
	}

	return skills, nil
}

func (r *Repository) SearchIntern(ctx context.Context, filters models.SearchIntern) ([]models.ShortProfile, error) {
	res := []models.ShortProfile{}

	var query strings.Builder
	query.WriteString("SELECT id, first_name, last_name, location, university, degree, image FROM interns")

	correctFl := false
	correct := func() {
		if !correctFl {
			query.WriteString(" WHERE ")
			correctFl = true
		} else {
			query.WriteString(" AND ")
		}
	}
	index := 1
	values := []interface{}{}
	if filters.Query != nil {
		correct()
		query.WriteString("tsv_content @@ to_tsquery('russian', $")
		query.WriteString(strconv.Itoa(index))
		query.WriteString(")")
		values = append(values, *filters.Query)
		index++
	}

	if filters.University != nil {
		correct()
		query.WriteString("university ILIKE $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, (fmt.Sprintf("%c%s%c", '%', *filters.University, '%')))
		index++
	}

	if filters.Query != nil {
		query.WriteString(" ORDER BY ts_rank(tsv_content, to_tsquery('russian', $1)) DESC")
	}

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
	}

	query.WriteString(";")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		intern := models.ShortProfile{}
		err = rows.Scan(
			&intern.ID,
			&intern.FirstName,
			&intern.LastName,
			&intern.Location,
			&intern.University,
			&intern.Degree,
			&intern.Image,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		res = append(res, intern)
	}
	return res, nil
}

func (r *Repository) GetUserIDByProfileID(ctx context.Context, id int) (int, error) {
	var userID int
	err := r.db.QueryRowContext(ctx, `SELECT user_id FROM interns WHERE id = $1`, id).Scan(&userID)
	if err != nil {
		return userID, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return userID, nil
}

func (r *Repository) ExistStatus(ctx context.Context, companyID, internID int, status string) (bool, error) {
	exist := false
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM applications 
		WHERE intern_id = $1 AND EXISTS(
			SELECT 1 FROM internships
			WHERE id = applications.internship_id AND company_id = $2) 
		AND status = $3)`, internID, companyID, status).Scan(&exist)
	if err != nil {
		return exist, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return exist, nil
}
