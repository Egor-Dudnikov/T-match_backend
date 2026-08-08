// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func (r *Repository) QueryProfile(ctx context.Context, id int, profile models.Profile) error {
	query := newUpdateQuery("UPDATE interns SET ", id)

	addFilled[string](query, "first_name", profile.FirstName)
	addFilled[string](query, "last_name", profile.LastName)
	addFilled[time.Time](query, "birth_date", profile.BirthDate)
	addFilled[string](query, "bio", profile.Bio)
	addFilled[string](query, "degree", profile.Degree)
	addFilled[string](query, "experience", profile.Experience)
	addFilled[string](query, "location", profile.Location)
	addFilled[string](query, "university", profile.University)

	if query.empty() {
		return nil
	}

	queryStr, values := query.parseBuilder(" WHERE user_id = $1")

	_, err := r.db.ExecContext(ctx, queryStr, values...)
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
	err := r.db.QueryRowContext(ctx, `
		SELECT i.id, i.user_id, i.first_name, i.last_name, i.birth_date, i.location, 
		       i.university, i.degree, i.bio, i.experience, i.image 
		FROM interns i
		WHERE i.id = $1 
		  AND NOT EXISTS(
			  SELECT 1 FROM user_bans ub 
			  WHERE ub.user_id = i.user_id
		  )
	`, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.BirthDate,
		&profile.Location,
		&profile.University,
		&profile.Degree,
		&profile.Bio,
		&profile.Experience,
		&profile.Image,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return profile, apierrors.ErrProfileNotFound
		}
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
			return skills, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}
		skills = append(skills, models.Skill{ID: id, Name: name})
	}

	return skills, nil
}

func (r *Repository) SearchIntern(ctx context.Context, filters models.SearchIntern) ([]models.ShortProfile, error) {
	res := []models.ShortProfile{}

	query := newQuerySelectBuilder(`
		SELECT i.id, i.user_id, i.first_name, i.last_name, i.location, i.university, i.degree, i.image 
		FROM interns i
	`)

	query.addWhere(`
		NOT EXISTS(
			SELECT 1 FROM user_bans ub 
			WHERE ub.user_id = i.user_id
		)
	`)

	addWhereWithIndex[string](query, "tsv_content @@ to_tsquery('russian', $", ")", filters.Query)

	if filters.University != nil {
		university := fmt.Sprintf("%c%s%c", '%', *filters.University, '%')
		addWhereWithIndex[string](query, "university ILIKE $", "", &university)
	}

	if filters.Skills != nil {
		addWhereWithIndexes(query, "(SELECT COUNT(DISTINCT skill_id) FROM intern_skills WHERE intern_id = i.id AND skill_id ",
			") = "+strconv.Itoa(len(*filters.Skills)), filters.Skills)
	}

	if filters.Query != nil {
		query.addOrderBy(" ts_rank(tsv_content, to_tsquery('russian', $1)) ", constants.DESC)
	}
	query.addLimit(filters.Limit)
	query.addOffset(filters.Offset)

	queryStr, values := query.parseBuilder()

	rows, err := r.db.QueryContext(ctx, queryStr, values...)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		intern := models.ShortProfile{}
		err = rows.Scan(
			&intern.ID,
			&intern.UserID,
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
