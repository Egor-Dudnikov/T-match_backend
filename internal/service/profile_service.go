// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"time"
)

func (app Service) UpdateStudentProfile(ctx context.Context, profile models.Profile) error {
	err := app.validate.Struct(profile)
	if profile.BirthDate != nil {
		now := time.Now()
		age := now.Year() - profile.BirthDate.Year()
		if profile.BirthDate.YearDay() < now.YearDay() {
			age--
		}
		if age < 16 {
			return apierrors.ErrUserMustBe16
		}
	}

	claims := ctx.Value("claims").(models.Claims)
	err = app.db.QueryProfile(ctx, claims.UserID, profile)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (app Service) GetMyProfile(ctx context.Context) (models.ProfileResponse, error) {
	claims := ctx.Value("claims").(models.Claims)
	resp := models.ProfileResponse{Email: claims.Email}
	profile, err := app.db.GetMyProfile(ctx, claims.UserID)
	resp.Profile = profile
	if err != nil {
		return resp, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	resp.Skills, err = app.db.GetInternSkills(ctx, claims.UserID)
	if err != nil {
		return resp, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
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
	claims := ctx.Value("claims").(models.Claims)
	resp := models.CompanyProfileResponse{
		Email: claims.Email}
	profile, err := app.db.GetCompanyProfile(ctx, claims.UserID)
	resp.Profile = profile
	if err != nil {
		return resp, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return resp, nil
}

func (app Service) SetMyAvatar(ctx context.Context, info *multipart.FileHeader, file io.Reader, claims models.Claims) (string, error) {

	name := "user:" + strconv.Itoa(claims.UserID) + ":avatar"
	url, err := app.s3.SetFile(ctx, name, file, "image/jpeg", info)
	if err != nil {
		return url, err
	}

	if claims.Role == "company" {
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
		return skills, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return skills, nil
}

func (app Service) AddInternSkills(ctx context.Context, skills []int) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.AddInternSkills(ctx, skills, claims.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (app Service) DeleteInternSkills(ctx context.Context, skillIDs []int) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.DeleteInternSkills(ctx, skillIDs, claims.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (app Service) GetMyResponses(ctx context.Context) ([]models.Response, error) {
	responses := []models.Response{}
	claims := ctx.Value("claims").(models.Claims)
	internID, err := app.db.GetInternId(ctx, claims.UserID)
	if err != nil {
		return responses, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	responses, err = app.db.GetMyResponses(ctx, internID)
	return responses, err
}

func (app Service) SearchCompany(ctx context.Context, filters models.SearchCompany) ([]models.Company, error) {
	res, err := app.db.SearchCompany(ctx, filters)
	return res, err
}
