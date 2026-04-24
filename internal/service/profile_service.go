package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"context"
	"fmt"
)

func (app AuthService) NewProfile(ctx context.Context, profile models.Profile) error {
	claims := ctx.Value("claims").(models.Claims)
	err := app.db.QueryProfile(ctx, claims.UserID, profile)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	return nil
}
