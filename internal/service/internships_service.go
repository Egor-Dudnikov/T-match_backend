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
