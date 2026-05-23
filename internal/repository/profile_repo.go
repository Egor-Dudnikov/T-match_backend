// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
	"T-match_backend/internal/models"
	"context"
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
		return err
	}
	return nil
}

func (r *Repository) GetMyProfile(ctx context.Context, id int) (models.Profile, error) {
	profile := models.Profile{}
	err := r.db.QueryRowContext(ctx, `SELECT first_name, last_name, birth_date, location, university, degree, bio, experience, image 
	FROM interns
	WHERE user_id = $1`, id).Scan(
		&profile.FirstName,
		&profile.LastName,
		&profile.BirthDate,
		&profile.Location,
		&profile.University,
		&profile.Degree,
		&profile.Bio,
		&profile.Experience,
		&profile.Image)
	return profile, err
}

func (r *Repository) UpdateCompanyProfile(ctx context.Context, id int, profile models.CompanyProfile) error {
	var query strings.Builder
	query.WriteString("UPDATE companies SET ")
	values := []interface{}{id}
	cnt := 1
	delimiter := false
	addfilled := func(filled string, value any) {
		cnt++
		if delimiter {
			query.WriteString(" ,")
		}
		query.WriteString(filled)
		query.WriteString(" = $")
		query.WriteString(strconv.Itoa(cnt))

		values = append(values, value)
		delimiter = true
	}

	if profile.CompanyName != nil {
		addfilled("company_name", *profile.CompanyName)
	}
	if profile.Description != nil {
		addfilled("description", *profile.Description)
	}
	if profile.Website != nil {
		addfilled("website", *profile.Website)
	}

	if cnt == 1 {
		return nil
	}

	query.WriteString(" WHERE user_id = $1")

	_, err := r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetCompanyProfile(ctx context.Context, id int) (models.CompanyProfile, error) {
	profile := models.CompanyProfile{}
	err := r.db.QueryRowContext(ctx, `SELECT company_name, description, website, inn, ogrn, kpp, legal_address, director_name, image
	FROM companies 
	WHERE user_id = $1`, id).Scan(
		&profile.CompanyName,
		&profile.Description,
		&profile.Website,
		&profile.Inn,
		&profile.Ogrn,
		&profile.Kpp,
		&profile.LegalAddress,
		&profile.DirectorName,
		&profile.Image)
	return profile, err
}

func (r *Repository) SetMyAvatar(ctx context.Context, url string, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE interns SET
	image = $2
	WHERE user_id = $1`, id, url)
	return err
}

func (r *Repository) SetMyCompanyAvatar(ctx context.Context, url string, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE companies SET
	image = $2
	WHERE user_id = $1`, id, url)
	return err
}

func (r *Repository) GetAllSkills(ctx context.Context) ([]models.Skill, error) {
	skills := []models.Skill{}
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM skills;`)
	if err != nil {
		return skills, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		skills = append(skills, models.Skill{Id: id, Name: name})
	}

	return skills, nil
}

func SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.CompanyProfile, error) {
	res := []models.CompanyProfile{}

	var query strings.Builder
	query.WriteString("SELECT id, user_id, company_name, description, website, inn, kpp, ogrn, legal_address, director_name, image FROM companies ")
	query.WriteString("WHERE ")

	values := []interface{}{}
	if filters.Query != nil {
		values = append(values, query)
	}
	return res, nil
}
