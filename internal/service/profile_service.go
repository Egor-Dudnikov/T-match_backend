// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"io"
	"mime/multipart"
	"strconv"
)

func (app Service) UpdateStudentProfile(ctx context.Context, profile models.Profile) error {
	err := app.validate.Struct(profile)
	if profile.BirthDate != nil {
		if !utils.ValidAge(*profile.BirthDate) {
			return apierrors.ErrUserMustBe16
		}
	}

	claims := ctx.Value("claims").(models.Claims)
	err = app.db.QueryProfile(ctx, claims.UserID, profile)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) GetMyProfile(ctx context.Context) (models.ProfileResponse, error) {
	claims := ctx.Value("claims").(models.Claims)
	resp := models.ProfileResponse{}
	id, err := app.db.GetProfileIdByUserId(ctx, claims.UserID)
	if err != nil {
		return resp, err
	}
	resp, err = app.profileResponse(ctx, id, claims.UserID, claims.Email)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app Service) UpdateCompanyProfile(ctx context.Context, profile models.CompanyProfile) error {
	err := app.validate.Struct(profile)
	if err != nil {
		return err
	}
	claims := ctx.Value("claims").(models.Claims)
	err = app.db.UpdateCompanyProfile(ctx, claims.UserID, profile)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) GetMyCompanyProfile(ctx context.Context) (models.CompanyProfileResponse, error) {
	resp := models.CompanyProfileResponse{}

	claims := ctx.Value("claims").(models.Claims)

	id, err := app.db.GetCompanyIdByUserId(ctx, claims.UserID)
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

func (app Service) GetCompanyProfile(ctx context.Context, id int) (models.CompanyProfileResponse, error) {
	resp := models.CompanyProfileResponse{}

	claims := ctx.Value("claims").(models.Claims)
	var email string

	if claims.Role == constants.Intern {

		internId, err := app.db.GetProfileIdByUserId(ctx, claims.UserID)
		if err != nil {
			return resp, err
		}

		exist, err := app.existStudentMatch(ctx, id, internId)
		if err != nil {
			return resp, err
		}
		if exist {
			userId, err := app.db.GetUserIdByCompanyId(ctx, id)
			if err != nil {
				return resp, err
			}
			email, err = app.db.GetEmailByUserId(ctx, userId)
			if err != nil {
				return resp, err
			}
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

func (app Service) SetMyAvatar(ctx context.Context, info *multipart.FileHeader, file io.Reader, claims models.Claims) (string, error) {
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
		app.s3.Delete(ctx, name)
		return "", err
	}
	return url, err
}

func (app Service) GetAllSkills(ctx context.Context) ([]models.Skill, error) {
	skills, err := app.db.GetAllSkills(ctx)
	if err != nil {
		return skills, err
	}
	return skills, nil
}

func (app Service) AddInternSkills(ctx context.Context, skills []int) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.AddInternSkills(ctx, skills, claims.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) DeleteInternSkills(ctx context.Context, skillIDs []int) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.DeleteInternSkills(ctx, skillIDs, claims.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) GetMyResponses(ctx context.Context) ([]models.Response, error) {
	responses := []models.Response{}
	claims := ctx.Value("claims").(models.Claims)
	internID, err := app.db.GetProfileIdByUserId(ctx, claims.UserID)
	if err != nil {
		return responses, err
	}
	responses, err = app.db.GetMyResponses(ctx, internID)
	return responses, err
}

func (app Service) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.CompanyProfile, error) {
	res, err := app.db.SearchCompany(ctx, filters)
	return res, err
}

func (app Service) SearchIntern(ctx context.Context, filters models.SearchIntern) ([]models.ShortProfile, error) {
	res, err := app.db.SearchIntern(ctx, filters)
	return res, err
}

func (app Service) GetProfile(ctx context.Context, id int) (models.ProfileResponse, error) {
	resp := models.ProfileResponse{}
	claims := ctx.Value("claims").(models.Claims)
	userId, err := app.db.GetUserIdByProfileId(ctx, id)
	var email string
	if claims.Role == constants.Company {
		companyId, err := app.db.GetCompanyIdByUserId(ctx, claims.UserID)
		if err != nil {
			return resp, err
		}
		exist, err := app.existCompanyMatch(ctx, companyId, id)
		if err != nil {
			return resp, err
		}
		if exist {
			email, err = app.db.GetEmailByUserId(ctx, userId)
			if err != nil {
				return resp, err
			}
		}
	}
	resp, err = app.profileResponse(ctx, id, userId, email)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app Service) profileResponse(ctx context.Context, internId int, userId int, email string) (models.ProfileResponse, error) {
	resp := models.ProfileResponse{Email: email}
	profile, err := app.db.GetProfile(ctx, internId)
	resp.Profile = profile
	if err != nil {
		return resp, err
	}
	resp.Skills, err = app.db.GetInternSkills(ctx, userId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (app Service) existStudentMatch(ctx context.Context, companyId, internId int) (bool, error) {
	exist, err := app.db.ExistStatus(ctx, companyId, internId, constants.Accepted)
	if err != nil {
		return exist, err
	}
	return exist, nil
}

func (app Service) existCompanyMatch(ctx context.Context, companyId, internId int) (bool, error) {
	exist, err := app.db.ExistStatus(ctx, companyId, internId, constants.Reviewing)
	if err != nil {
		return exist, err
	}
	return exist, nil
}
