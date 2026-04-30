package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
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

func (app Service) GetMyProfile(ctx context.Context) (models.Profile, error) {
	claims := ctx.Value("claims").(models.Claims)
	profile, err := app.db.GetMyProfile(ctx, claims.UserID)
	if err != nil {
		return profile, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return profile, nil
}

func (app Service) UpdateCompanyProfile(ctx context.Context, profile models.CompanyProfile) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.UpdateCompanyProfile(ctx, claims.UserID, profile)
	if err != nil {
		return err
	}
	return nil
}

func (app Service) GetMyCompanyProfile(ctx context.Context) (models.CompanyProfile, error) {
	claims := ctx.Value("claims").(models.Claims)
	profile, err := app.db.GetCompanyProfile(ctx, claims.UserID)
	if err != nil {
		return profile, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return profile, nil
}
