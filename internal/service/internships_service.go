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
		return 0, apierrors.Warp(apierrors.ErrBadRequest, err)
	}
	internshipID, err := app.db.NewInternship(ctx, internship, id)
	if err != nil {
		return internshipID, err
	}
	return internshipID, nil
}

func (app Service) GetInternshipById(ctx context.Context, id int) (models.InternshipResponse, error) {
	res := models.InternshipResponse{}
	internship, err := app.db.GetInternshipById(ctx, id)
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
		return apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	err = app.IsCompanysInternship(ctx, internship.Id)
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
	companyId, err := app.db.GetCompanyIdByInternshipId(ctx, id)
	if err != nil {
		return err
	}

	companyIdUser, err := app.db.GetCompanyIdByUserId(ctx, ctx.Value("claims").(models.Claims).UserID)
	if companyId != companyIdUser {
		return apierrors.ErrForbidden
	}
	return nil
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
	claims := ctx.Value("claims").(models.Claims)
	internID, err := app.db.GetProfileIdByUserId(ctx, claims.UserID)
	if err != nil {
		return err
	}
	err = app.db.RespondInternship(ctx, internID, internshipID)
	if err != nil {
		return err
	}
	return nil
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
	internshipID, err := app.db.GetInternshipIdByResponseId(ctx, responseID)
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
