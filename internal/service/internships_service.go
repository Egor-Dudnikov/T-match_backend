package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
)

func (app Service) NewInternship(ctx context.Context, internship models.Internship, id int) (int, error) {
	err := app.validate.Struct(internship)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}
	internshipID, err := app.db.NewInternship(ctx, internship, id)
	if err != nil {
		return internshipID, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return internshipID, nil
}

func (app Service) GetInternshipById(ctx context.Context, id int) (models.InternshipResponse, error) {
	res := models.InternshipResponse{}
	internship, err := app.db.GetInternshipById(ctx, id)
	if err != nil {
		return res, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	res.Internship = internship
	res.Skills, err = app.db.GetInternshipSkills(ctx, id)
	if err != nil {
		return res, fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	return res, nil
}

func (app Service) UpdateInternship(ctx context.Context, internship models.InternshipUpdate) error {
	err := app.validate.Struct(internship)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
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
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}

func (app Service) IsCompanysInternship(ctx context.Context, id int) error {
	companyId, err := app.db.GetCompanyIdByInternshipId(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	companyIdUser, err := app.db.GetCmpanyIdByUserId(ctx, ctx.Value("claims").(models.Claims).UserID)
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
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
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
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}
