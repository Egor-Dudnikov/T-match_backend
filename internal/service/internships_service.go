// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
)

func (app Service) NewInternship(ctx context.Context, internship models.Internship, id int) (int, error) {
	err := app.validate.Struct(internship)
	if err != nil {
		return 0, apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	internshipID, err := app.db.NewInternship(ctx, internship, id)
	if err != nil {
		return internshipID, err
	}
	return internshipID, nil
}

func (app Service) GetInternshipByID(ctx context.Context, id int) (models.InternshipResponse, error) {
	res := models.InternshipResponse{}
	internship, err := app.db.GetInternshipByID(ctx, id)
	if err != nil {
		return res, err
	}

	res.Internship = internship
	res.Skills, err = app.db.GetInternshipSkills(ctx, id)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (app Service) UpdateInternship(ctx context.Context, internship models.InternshipUpdate) error {
	err := app.validate.Struct(internship)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	err = app.IsCompanysInternship(ctx, internship.ID)
	if err != nil {
		return err
	}

	err = app.db.UpdateInternships(ctx, internship)
	return err
}

func (app Service) ArchivedInternship(ctx context.Context, id int) error {
	err := app.IsCompanysInternship(ctx, id)
	if err != nil {
		return err
	}
	err = app.db.ArchivedInternship(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) IsCompanysInternship(ctx context.Context, id int) error {
	companyID, err := app.db.GetCompanyIDByInternshipID(ctx, id)
	if err != nil {
		return err
	}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	companyIDUser, err := app.db.GetCompanyIDByUserID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if companyID != companyIDUser {
		return apierrors.ErrForbidden
	}
	return nil
}

func (app Service) GetCompanyesInternshipsByUserID(ctx context.Context, userID int) ([]models.Internship, error) {
	res := []models.Internship{}

	id, err := app.db.GetCompanyIDByUserID(ctx, userID)
	if err != nil {
		return res, err
	}

	res, err = app.db.GetCompanyInternships(ctx, id, false)
	return res, err
}

func (app Service) GetCompanyesInternships(ctx context.Context, companyID int) ([]models.Internship, error) {
	res, err := app.db.GetCompanyInternships(ctx, companyID, true)
	return res, err
}

func (app Service) AddInternshipSkills(ctx context.Context, skills []int, id int) error {
	err := app.IsCompanysInternship(ctx, id)
	if err != nil {
		return err
	}
	err = app.db.AddInternshipSkills(ctx, skills, id)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) DeleteInternshipSkills(ctx context.Context, internshipID int, skillIDs []int) error {
	err := app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return err
	}
	err = app.db.DeleteInternshipSkills(ctx, skillIDs, internshipID)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) RespondInternship(ctx context.Context, internshipID int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	internID, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	err = app.db.RespondInternship(ctx, internID, internshipID)
	return err
}

func (app Service) DeleteRespondInternship(ctx context.Context, internshipID int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	internID, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	err = app.db.DeleteRespondInternship(ctx, internID, internshipID)
	return err
}

func (app Service) GetInternshipResponses(ctx context.Context, internshipID int) ([]models.Response, error) {
	err := app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return []models.Response{}, err
	}
	responses, err := app.db.InternshipsResponse(ctx, internshipID)
	return responses, err
}

func (app Service) SetResponseStatus(ctx context.Context, responseID int, status string) error {
	statuses := [4]string{constants.Pending, constants.Reviewing, constants.Accepted, constants.Rejected}
	ok := false
	for _, st := range statuses {
		if status == st {
			ok = true
		}
	}
	if !ok {
		return apierrors.ErrBadRequest
	}
	internshipID, err := app.db.GetInternshipIDByResponseID(ctx, responseID)
	if err != nil {
		return err
	}

	err = app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return err
	}

	err = app.db.SetResponseStatus(ctx, responseID, status)
	return err
}

func (app Service) SearchInternship(ctx context.Context, filters models.SearchInternship) ([]models.Internship, error) {
	res, err := app.db.SearchInternship(ctx, filters)
	return res, err
}
