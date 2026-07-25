package repository

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"fmt"
)

func (r *Repository) UpdateCompanyProfile(ctx context.Context, id int, profile models.CompanyProfile) error {
	query := newUpdateQuery("UPDATE companies SET ", id)

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
	err := r.db.QueryRowContext(ctx, `SELECT id, company_name, description, website, inn, ogrn, kpp, legal_address, director_name, image
	FROM companies 
	WHERE id = $1`, id).Scan(
		&profile.ID,
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

	query := newQuerySelectBuilder("SELECT id, company_name, description, website, inn, ogrn, legal_address, image FROM companies ")

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
