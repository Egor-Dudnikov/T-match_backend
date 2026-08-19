// Package service implements the business logic layer of the backend.
package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"log"
)

// GetAdminStats returns the admin statistics with the current number of online
// users.
func (app *Service) GetAdminStats(ctx context.Context) (models.AdminStats, error) {
	stats, err := app.db.GetStats(ctx)
	if err != nil {
		return stats, err
	}

	stats.UsersOnline = app.Hub.GetOnlineCount()

	return stats, nil
}

// BanUser bans the user with the given ID and kicks them out of the hub.
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

	role, err := app.db.GetUserRole(ctx, userID)
	if err != nil {
		return err
	}
	if role == constants.Admin {
		return apierrors.ErrCannotBanAdmin
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

	if err := app.cache.DeleteUserSessions(ctx, userID); err != nil {
		log.Printf("failed to delete sessions of banned user %d: %v", userID, err)
	}

	return nil
}

// UnbanUser removes the ban from the user with the given ID.
func (app *Service) UnbanUser(ctx context.Context, userID int) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	if claims.UserID == userID {
		return apierrors.ErrBadRequest
	}

	if _, err := app.db.GetUserRole(ctx, userID); err != nil {
		return err
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

// IsUserBanned reports whether the user with the given ID is banned.
func (app *Service) IsUserBanned(ctx context.Context, userID int) (bool, error) {
	return app.db.IsUserBanned(ctx, userID)
}

// GetUserBanReason returns the reason the user with the given ID was banned.
func (app *Service) GetUserBanReason(ctx context.Context, userID int) (string, error) {
	return app.db.GetUserBanReason(ctx, userID)
}

// HandlingBannedUser returns the ban reason if the user is banned, or an empty
// string otherwise.
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

// AdminDeleteInternship deletes the internship with the given ID.
func (app *Service) AdminDeleteInternship(ctx context.Context, internshipID int) error {
	exists, err := app.db.InternshipExists(ctx, internshipID)
	if err != nil {
		return err
	}
	if !exists {
		return apierrors.ErrInternshipNotFound
	}

	err = app.db.AdminDeleteInternship(ctx, internshipID)
	if err != nil {
		return err
	}

	app.deleteRecsysInternship(ctx, internshipID)

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	log.Printf("Admin %d deleted internship %d", claims.UserID, internshipID)

	return nil
}
