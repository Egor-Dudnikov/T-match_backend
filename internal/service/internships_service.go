package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
)

func (app Service) NewInternship(ctx context.Context, internship models.Internship, id int) error {
	err := app.validate.Struct(internship)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}
	err = app.db.NewInternship(ctx, internship, id)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (app Service) GetInternshipById(ctx context.Context, id int) (models.Internship, error) {
	internship, err := app.db.GetInternshipById(ctx, id)
	if err != nil {
		return internship, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	if internship.IsArchived {
		claims := ctx.Value("claims").(models.Claims)
		CompanyId, err := app.db.GetCmpanyIdByUserId(ctx, claims.UserID)
		if err != nil {
			return internship, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
		}
		if CompanyId != internship.CompanyId {
			return internship, apierrors.ErrInternshipIsArchived
		}
	}
	return internship, err
}
