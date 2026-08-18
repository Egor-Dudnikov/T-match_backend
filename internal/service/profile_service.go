// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
)

func (app *Service) UpdateStudentProfile(ctx context.Context, profile models.Profile) error {
	err := app.validate.Struct(profile)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	if profile.BirthDate != nil {
		if !utils.ValidAge(*profile.BirthDate) {
			return apierrors.ErrUserMustBe16
		}
	}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	err = app.db.QueryProfile(ctx, claims.UserID, profile)
	if err != nil {
		return err
	}
	if profile.CityID != nil {
		app.syncUserGeo(ctx, claims.UserID, *profile.CityID)
	}
	return nil
}

func (app *Service) GetMyProfile(ctx context.Context) (models.ProfileResponse, error) {
	resp := models.ProfileResponse{}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return resp, apierrors.ErrInternalServer
	}

	id, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
	if err != nil {
		return resp, err
	}
	resp, err = app.profileResponse(ctx, id, claims.UserID, claims.Email)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app *Service) UpdateCompanyProfile(ctx context.Context, profile models.CompanyProfile) error {
	err := app.validate.Struct(profile)
	if err != nil {
		return err
	}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	err = app.db.UpdateCompanyProfile(ctx, claims.UserID, profile)
	if err != nil {
		return err
	}
	return nil
}

func (app *Service) GetMyCompanyProfile(ctx context.Context) (models.CompanyProfileResponse, error) {
	resp := models.CompanyProfileResponse{}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return resp, apierrors.ErrInternalServer
	}

	id, err := app.db.GetCompanyIDByUserID(ctx, claims.UserID)
	if err != nil {
		return resp, err
	}

	profile, err := app.db.GetCompanyProfile(ctx, id)
	resp.Profile = profile
	if err != nil {
		return resp, err
	}
	resp.Email = claims.Email
	return resp, nil
}

func (app *Service) GetCompanyProfile(ctx context.Context, id int) (models.CompanyProfileResponse, error) {
	resp := models.CompanyProfileResponse{}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)

	var email string
	var existMatch bool

	var err error

	if ok && claims.Role == constants.Intern {
		internID, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
		if err != nil {
			return resp, err
		}

		existMatch, err = app.existStudentMatch(ctx, id, internID)
		if err != nil {
			return resp, err
		}
	}

	if ok && (existMatch || claims.Role == constants.Admin) {
		email, err = app.companyEmail(ctx, id)
		if err != nil {
			return resp, err
		}
	}

	resp.Email = email
	profile, err := app.db.GetCompanyProfile(ctx, id)
	resp.Profile = profile
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app *Service) companyEmail(ctx context.Context, companyID int) (string, error) {
	userID, err := app.db.GetUserIDByCompanyID(ctx, companyID)
	if err != nil {
		return "", err
	}
	return app.db.GetEmailByUserID(ctx, userID)
}

func (app *Service) SetMyAvatar(ctx context.Context, info *multipart.FileHeader, file io.Reader, claims models.Claims) (string, error) {
	if info.Size > constants.MaxSizeImage {
		return "", apierrors.ErrBadRequest
	}

	contentType := info.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		return "", apierrors.ErrBadRequest
	}

	name := "user:" + strconv.Itoa(claims.UserID) + ":avatar"
	url, err := app.s3.SetFile(ctx, name, file, contentType, info)
	if err != nil {
		return url, err
	}

	if claims.Role == constants.Company {
		err = app.db.SetMyCompanyAvatar(ctx, url, claims.UserID)
	} else {
		err = app.db.SetMyAvatar(ctx, url, claims.UserID)
	}

	if err != nil {
		s3Err := app.s3.Delete(ctx, name)
		if s3Err != nil {
			return "", fmt.Errorf("s3 delete error: %w, original error: %w", s3Err, err)
		}
		return "", err
	}
	return url, err
}

func (app *Service) GetAllSkills(ctx context.Context) ([]models.Skill, error) {
	skills, err := app.db.GetAllSkills(ctx)
	if err != nil {
		return skills, err
	}
	return skills, nil
}

func (app *Service) GetAllCities(ctx context.Context) ([]models.City, error) {
	cities, err := app.db.GetAllCities(ctx)
	if err != nil {
		return cities, err
	}
	return cities, nil
}

func (app *Service) AddInternSkills(ctx context.Context, skills []int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	err := app.db.AddInternSkills(ctx, skills, claims.UserID)
	if err != nil {
		return err
	}
	app.addRecsysUserSkills(ctx, claims.UserID, skills)
	return nil
}

func (app *Service) DeleteInternSkills(ctx context.Context, skillIDs []int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	err := app.db.DeleteInternSkills(ctx, skillIDs, claims.UserID)
	if err != nil {
		return err
	}
	app.deleteRecsysUserSkills(ctx, claims.UserID, skillIDs)
	return nil
}

func (app *Service) GetMyResponses(ctx context.Context) ([]models.Response, error) {
	responses := []models.Response{}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return responses, apierrors.ErrInternalServer
	}

	internID, err := app.db.GetProfileIDByUserID(ctx, claims.UserID)
	if err != nil {
		return responses, err
	}

	responses, err = app.db.GetMyResponses(ctx, internID)
	return responses, err
}

func (app *Service) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.CompanyProfile, error) {
	res, err := app.db.SearchCompany(ctx, filters)
	return res, err
}

func (app *Service) SearchIntern(ctx context.Context, filters models.SearchIntern) ([]models.ShortProfile, error) {
	res, err := app.db.SearchIntern(ctx, filters)
	return res, err
}

func (app *Service) GetProfile(ctx context.Context, internID int) (models.ProfileResponse, error) {
	resp := models.ProfileResponse{}
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)

	userID, err := app.db.GetUserIDByProfileID(ctx, internID)
	if err != nil {
		return resp, err
	}

	var email string
	var existMatch bool

	if ok && claims.Role == constants.Company {
		existMatch, err = app.companyMatchesIntern(ctx, claims, internID)
		if err != nil {
			return resp, err
		}
	}

	if ok && (existMatch || claims.Role == constants.Admin) {
		email, err = app.db.GetEmailByUserID(ctx, userID)
		if err != nil {
			return resp, err
		}
	}

	resp, err = app.profileResponse(ctx, internID, userID, email)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app *Service) companyMatchesIntern(ctx context.Context, claims models.Claims, internID int) (bool, error) {
	companyID, err := app.db.GetCompanyIDByUserID(ctx, claims.UserID)
	if err != nil {
		return false, err
	}
	return app.existCompanyMatch(ctx, companyID, internID)
}

func (app *Service) profileResponse(ctx context.Context, internID int, userID int, email string) (models.ProfileResponse, error) {
	resp := models.ProfileResponse{Email: email}
	profile, err := app.db.GetProfile(ctx, internID)
	resp.Profile = profile
	if err != nil {
		return resp, err
	}
	resp.Skills, err = app.db.GetInternSkills(ctx, userID)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app *Service) existStudentMatch(ctx context.Context, companyID, internID int) (bool, error) {
	exist, err := app.db.ExistStatus(ctx, companyID, internID, constants.Accepted)
	if err != nil {
		return exist, err
	}
	return exist, nil
}

func (app *Service) existCompanyMatch(ctx context.Context, companyID, internID int) (bool, error) {
	exist1, err := app.db.ExistStatus(ctx, companyID, internID, constants.Reviewing)
	if err != nil {
		return exist1, err
	}
	exist2, err := app.db.ExistStatus(ctx, companyID, internID, constants.Accepted)
	if err != nil {
		return exist2, err
	}
	return exist1 || exist2, nil
}
