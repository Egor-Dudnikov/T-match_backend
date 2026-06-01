package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
	"strconv"
	"strings"
)

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
	if err != nil {
		return apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) GetCompanyProfile(ctx context.Context, id int) (models.CompanyProfile, error) {
	profile := models.CompanyProfile{}
	err := r.db.QueryRowContext(ctx, `SELECT id, company_name, description, website, inn, ogrn, kpp, legal_address, director_name, image
	FROM companies 
	WHERE id = $1`, id).Scan(
		&profile.Id,
		&profile.CompanyName,
		&profile.Description,
		&profile.Website,
		&profile.Inn,
		&profile.Ogrn,
		&profile.Kpp,
		&profile.LegalAddress,
		&profile.DirectorName,
		&profile.Image)
	if err != nil {
		return profile, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return profile, nil
}

func (r *Repository) SetMyCompanyAvatar(ctx context.Context, url string, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE companies SET
	image = $2
	WHERE user_id = $1`, id, url)
	if err != nil {
		return apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.CompanyProfile, error) {
	res := []models.CompanyProfile{}

	var query strings.Builder
	query.WriteString("SELECT id, company_name, description, website, inn, ogrn, legal_address, image FROM companies ")

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
		query.WriteString("tsv_content @@ to_tsquery('russian', $")
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
		index++
	}

	query.WriteString(";")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, apierrors.Warp(apierrors.ErrDatabaseError, err)
	}
	defer rows.Close()

	for rows.Next() {
		company := models.CompanyProfile{}
		rows.Scan(
			&company.Id,
			&company.CompanyName,
			&company.Description,
			&company.Website,
			&company.Inn,
			&company.Ogrn,
			&company.LegalAddress,
			&company.Image,
		)
		res = append(res, company)
	}
	return res, nil
}
