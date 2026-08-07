package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (r *Repository) UpdateCompanyProfile(ctx context.Context, userID int, profile models.CompanyProfile) error {
	query := newUpdateQuery("UPDATE companies SET ", userID)

	addFilled[string](query, "company_name", profile.CompanyName)
	addFilled[string](query, "description", profile.Description)
	addFilled[string](query, "website", profile.Website)

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

func (r *Repository) GetCompanyProfile(ctx context.Context, id int) (models.CompanyProfile, error) {
	profile := models.CompanyProfile{}
	err := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.company_name, c.description, c.website, c.inn, c.ogrn, c.kpp, 
		       c.legal_address, c.director_name, c.image
		FROM companies c
		WHERE c.id = $1 
		  AND NOT EXISTS(
			  SELECT 1 FROM user_bans ub 
			  WHERE ub.user_id = c.user_id
		  )
	`, id).Scan(
		&profile.ID,
		&profile.CompanyName,
		&profile.Description,
		&profile.Website,
		&profile.Inn,
		&profile.Ogrn,
		&profile.Kpp,
		&profile.LegalAddress,
		&profile.DirectorName,
		&profile.Image,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return profile, apierrors.ErrCompanyNotFound
		}
		return profile, apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return profile, nil
}

func (r *Repository) SetMyCompanyAvatar(ctx context.Context, url string, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE companies SET
	image = $2
	WHERE user_id = $1`, id, url)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *Repository) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.CompanyProfile, error) {
	res := []models.CompanyProfile{}

	query := newQuerySelectBuilder(`
		SELECT c.id, c.company_name, c.description, c.website, c.inn, c.ogrn, c.legal_address, c.image 
		FROM companies c
	`)

	query.addWhere(`
		NOT EXISTS(
			SELECT 1 FROM user_bans ub 
			WHERE ub.user_id = c.user_id
		)
	`)

	addWhereWithIndex[string](query, "tsv_content @@ to_tsquery('russian', $", ")", filters.Query)

	if filters.Location != nil {
		location := fmt.Sprintf("%c%s%c", '%', *filters.Location, '%')
		addWhereWithIndex[string](query, "legal_address ILIKE $", "", &location)
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
		company := models.CompanyProfile{}
		err = rows.Scan(
			&company.ID,
			&company.CompanyName,
			&company.Description,
			&company.Website,
			&company.Inn,
			&company.Ogrn,
			&company.LegalAddress,
			&company.Image,
		)
		if err != nil {
			return res, apierrors.Wrap(apierrors.ErrDatabaseError, err)
		}

		res = append(res, company)
	}
	return res, nil
}
