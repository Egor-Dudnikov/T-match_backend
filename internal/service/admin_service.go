package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"log"
)

func (app *Service) GetAdminStats(ctx context.Context) (models.AdminStats, error) {
	stats, err := app.db.GetStats(ctx)
	if err != nil {
		return stats, err
	}

	stats.UsersOnline = app.Hub.GetOnlineCount()

	return stats, nil
}

func (app *Service) BanUser(ctx context.Context, userID int, adminBanRequest models.AdminBanRequest) error {
	err := app.validate.Struct(adminBanRequest)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	if claims.UserID == userID {
		return apierrors.ErrBadRequest
	}

	isBanned, err := app.db.IsUserBanned(ctx, userID)
	if err != nil {
		return err
	}
	if isBanned {
		return apierrors.ErrUserAlreadyBanned
	}

	err = app.db.BanUser(ctx, userID, claims.UserID, adminBanRequest.Reason)
	if err != nil {
		return err
	}

	app.Hub.KickUser(userID, adminBanRequest.Reason)

	return nil
}

func (app *Service) UnbanUser(ctx context.Context, userID int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	if claims.UserID == userID {
		return apierrors.ErrBadRequest
	}

	isBanned, err := app.db.IsUserBanned(ctx, userID)
	if err != nil {
		return err
	}
	if !isBanned {
		return apierrors.ErrUserNotBanned
	}

	return app.db.UnbanUser(ctx, userID)
}

func (app *Service) IsUserBanned(ctx context.Context, userID int) (bool, error) {
	return app.db.IsUserBanned(ctx, userID)
}

func (app *Service) GetUserBanReason(ctx context.Context, userID int) (string, error) {
	return app.db.GetUserBanReason(ctx, userID)
}

func (app *Service) HandlingBannedUser(ctx context.Context, userID int) (string, error) {
	isBanned, err := app.IsUserBanned(ctx, userID)
	if err != nil {
		return "", err
	}
	if isBanned {
		reason, err := app.GetUserBanReason(ctx, userID)
		if err != nil {
			return reason, err
		}
		return reason, apierrors.ErrUserBanned
	}
	return "", nil
}

func (app *Service) AdminDeleteInternship(ctx context.Context, internshipID int) error {
	err := app.db.AdminDeleteInternship(ctx, internshipID)
	if err != nil {
		return err
	}

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	log.Printf("Admin %d deleted internship %d", claims.UserID, internshipID)

	return nil
}
