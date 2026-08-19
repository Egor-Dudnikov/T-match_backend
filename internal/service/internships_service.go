// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"encoding/json"
)

// NewInternship validates and creates a new internship for the given company and returns its ID.
func (app *Service) NewInternship(ctx context.Context, internship models.Internship, id int) (int, error) {
	err := app.validate.Struct(internship)
	if err != nil {
		return 0, apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	internshipID, err := app.db.NewInternship(ctx, internship, id)
	if err != nil {
		return internshipID, err
	}
	app.syncInternshipCreate(ctx, internshipID, internship.CityID)
	return internshipID, nil
}

// GetInternshipByID returns the internship with the given ID, including its skills.
func (app *Service) GetInternshipByID(ctx context.Context, id int) (models.InternshipResponse, error) {
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

// UpdateInternship validates and updates an internship owned by the authenticated company.
func (app *Service) UpdateInternship(ctx context.Context, internship models.InternshipUpdate) error {
	err := app.validate.Struct(internship)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	err = app.IsCompanysInternship(ctx, internship.ID)
	if err != nil {
		return err
	}

	err = app.db.UpdateInternships(ctx, internship)
	if err != nil {
		return err
	}
	if internship.CityID != nil {
		app.syncInternshipGeo(ctx, internship.ID, *internship.CityID)
	}
	return nil
}

// ArchivedInternship archives the internship with the given ID.
func (app *Service) ArchivedInternship(ctx context.Context, id int) error {
	err := app.IsCompanysInternship(ctx, id)
	if err != nil {
		return err
	}
	err = app.db.ArchivedInternship(ctx, id)
	if err != nil {
		return err
	}
	app.deleteRecsysInternship(ctx, id)
	return nil
}

// IsCompanysInternship returns an error if the internship with the given ID does not belong to the authenticated company.
func (app *Service) IsCompanysInternship(ctx context.Context, id int) error {
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

// GetCompanyesInternshipsByUserID returns the non-archived internships of the company of the given user.
func (app *Service) GetCompanyesInternshipsByUserID(ctx context.Context, userID int) ([]models.Internship, error) {
	res := []models.Internship{}

	id, err := app.db.GetCompanyIDByUserID(ctx, userID)
	if err != nil {
		return res, err
	}

	res, err = app.db.GetCompanyInternships(ctx, id, false)
	return res, err
}

// GetCompanyesInternships returns the internships of the company with the given ID.
func (app *Service) GetCompanyesInternships(ctx context.Context, companyID int) ([]models.Internship, error) {
	res, err := app.db.GetCompanyInternships(ctx, companyID, true)
	return res, err
}

// AddInternshipSkills adds the given skills to the internship with the given ID.
func (app *Service) AddInternshipSkills(ctx context.Context, skills []int, id int) error {
	err := app.IsCompanysInternship(ctx, id)
	if err != nil {
		return err
	}
	err = app.db.AddInternshipSkills(ctx, skills, id)
	if err != nil {
		return err
	}
	app.addRecsysInternshipSkills(ctx, id, skills)
	return nil
}

// DeleteInternshipSkills deletes the given skills from the internship with the given ID.
func (app *Service) DeleteInternshipSkills(ctx context.Context, internshipID int, skillIDs []int) error {
	err := app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return err
	}

	err = app.db.DeleteInternshipSkills(ctx, skillIDs, internshipID)
	if err != nil {
		return err
	}

	app.deleteRecsysInternshipSkills(ctx, internshipID, skillIDs)

	return nil
}

// RespondInternship creates a response to the internship for the authenticated intern and notifies the company.
func (app *Service) RespondInternship(ctx context.Context, internshipID int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	internID, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
	if err != nil {
		return err
	}

	respondID, err := app.db.RespondInternship(ctx, internID, internshipID)
	if err != nil {
		return nil
	}

	app.sendRecsysAction(ctx, claims.UserID, internshipID, constants.RecsysActionApply)

	notification, err := app.db.NewApplicationNotification(ctx, internID, internshipID, respondID)
	if err != nil {
		return err
	}

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	app.Hub.Send(notification.UserID, string(notificationJSON))

	return err
}

// DeleteRespondInternship deletes the response of the authenticated intern to the internship with the given ID.
func (app *Service) DeleteRespondInternship(ctx context.Context, internshipID int) error {
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

// GetInternshipResponses returns the responses for the internship with the given ID.
func (app *Service) GetInternshipResponses(ctx context.Context, internshipID int) ([]models.Response, error) {
	err := app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return []models.Response{}, err
	}
	responses, err := app.db.InternshipsResponse(ctx, internshipID)
	return responses, err
}

// SetResponseStatus updates the status of the response with the given ID and notifies the intern.
func (app *Service) SetResponseStatus(ctx context.Context, responseID int, respReq models.ResponseRequest) error {
	err := app.validate.Struct(respReq)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	internshipID, err := app.db.GetInternshipIDByResponseID(ctx, responseID)
	if err != nil {
		return err
	}

	err = app.IsCompanysInternship(ctx, internshipID)
	if err != nil {
		return err
	}

	err = app.db.SetResponseStatus(ctx, responseID, respReq.Status)
	if err != nil {
		return err
	}

	notification, err := app.db.NewChangeStatusNotification(ctx, responseID, internshipID, respReq.Status)
	if err != nil {
		return err
	}

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	app.Hub.Send(notification.UserID, string(notificationJSON))

	return nil
}

// SearchInternship searches internships by the given filters.
func (app *Service) SearchInternship(ctx context.Context, filters models.SearchInternship) ([]models.Internship, error) {
	res, err := app.db.SearchInternship(ctx, filters)
	return res, err
}
