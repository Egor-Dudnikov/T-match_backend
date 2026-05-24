// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package repository

import (
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

func (r *Repository) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.Company, error) {
	res := []models.Company{}

	var query strings.Builder
	query.WriteString("SELECT id, company_name, description, website, inn, kpp, ogrn, legal_address, director_name, image FROM companies ")

	correctFl := false
	correct := func() {
		if !correctFl {
			query.WriteString("WHERE ")
			correctFl = true
		} else {
			query.WriteString(" AND ")
		}
	}
	index := 1
	values := []interface{}{}
	if filters.Query != nil {
		correct()
		query.WriteString("tsv_content @@ plainto_tsquery('russian', $")
		query.WriteString(strconv.Itoa(index))
		query.WriteString(")")
		values = append(values, *filters.Query)
		index++
	}

	if filters.Location != nil {
		correct()
		query.WriteString("legal_address ILIKE $")
		query.WriteString(strconv.Itoa(index))
		values = append(values, (fmt.Sprintf("%c%s%c", '%', *filters.Location, '%')))
		index++
	}

	if filters.Query != nil {
		query.WriteString("ORDER BY ts_rank(tsv_content, plainto_tsquery('russian', $1)) DESC")
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
		index++
	}

	query.WriteString(";")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		company := models.Company{}
		rows.Scan(
			&company.Id,
			&company.CompanyName,
			&company.Description,
			&company.Website,
			&company.Inn,
			&company.Kpp,
			&company.Ogrn,
			&company.LegalAddress,
			&company.DirectorName,
			&company.Image,
		)
		res = append(res, company)
	}
	return res, nil
}

func (r *Repository) SearchIntern(ctx context.Context, filters models.SearchIntern) ([]models.Intern, error) {
	res := []models.Intern{}

	var query strings.Builder
	query.WriteString("SELECT id, first_name, last_name, birth_date, location, university, degree, bio, experience, image FROM interns")

	correctFl := false
	correct := func() {
		if !correctFl {
			query.WriteString("WHERE ")
			correctFl = true
		} else {
			query.WriteString(" AND ")
		}
	}
	index := 1
	values := []interface{}{}
	if filters.Query != nil {
		correct()
		query.WriteString("tsv_content @@ plainto_tsquery('russian', $")
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
		query.WriteString(" ORDER BY ts_rank(tsv_content, plainto_tsquery('russian', $1)) DESC")
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
		index++
	}

	query.WriteString(";")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		intern := models.Intern{}
		rows.Scan(
			&intern.Id,
			&intern.FirstName,
			&intern.LastName,
			&intern.BirthDate,
			&intern.Location,
			&intern.University,
			&intern.Degree,
			&intern.Bio,
			&intern.Experience,
			&intern.Image,
		)
		res = append(res, intern)
	}
	return res, nil
}
