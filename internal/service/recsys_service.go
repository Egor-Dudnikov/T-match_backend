// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"context"
	"log"
)

// createRecsysUser registers a new intern in the recommendation service.
func (app *Service) createRecsysUser(ctx context.Context, userID int) {
	err := app.recsys.CreateUser(ctx, userID)
	if err != nil {
		log.Printf("recsys: create user %d: %v", userID, err)
	}
}

// syncUserGeo pushes the intern geo coordinates (mapped from city_id) to recsys.
func (app *Service) syncUserGeo(ctx context.Context, userID int, cityID int) {
	geoLen, geoLot, err := app.db.GetCityGeo(ctx, cityID)
	if err != nil {
		log.Printf("recsys: city geo for user %d: %v", userID, err)
		return
	}
	err = app.recsys.UpdateUserGeo(ctx, userID, geoLen, geoLot)
	if err != nil {
		log.Printf("recsys: update user %d geo: %v", userID, err)
	}
}

// syncInternshipCreate registers a new internship in recsys.
func (app *Service) syncInternshipCreate(ctx context.Context, internshipID int, cityID int) {
	geoLen, geoLot, err := app.db.GetCityGeo(ctx, cityID)
	if err != nil {
		log.Printf("recsys: city geo for internship %d: %v", internshipID, err)
		return
	}
	err = app.recsys.CreateInternship(ctx, internshipID, geoLen, geoLot)
	if err != nil {
		log.Printf("recsys: create internship %d: %v", internshipID, err)
	}
}

// syncInternshipGeo pushes the internship geo (mapped from city_id) to recsys.
func (app *Service) syncInternshipGeo(ctx context.Context, internshipID int, cityID int) {
	geoLen, geoLot, err := app.db.GetCityGeo(ctx, cityID)
	if err != nil {
		log.Printf("recsys: city geo for internship %d: %v", internshipID, err)
		return
	}
	err = app.recsys.UpdateInternshipGeo(ctx, internshipID, geoLen, geoLot)
	if err != nil {
		log.Printf("recsys: update internship %d geo: %v", internshipID, err)
	}
}

func (app *Service) deleteRecsysInternship(ctx context.Context, internshipID int) {
	err := app.recsys.DeleteInternship(ctx, internshipID)
	if err != nil {
		log.Printf("recsys: delete internship %d: %v", internshipID, err)
	}
}

func (app *Service) addRecsysUserSkills(ctx context.Context, userID int, skillIDs []int) {
	for _, skillID := range skillIDs {
		if err := app.recsys.AddUserSkill(ctx, userID, skillID); err != nil {
			log.Printf("recsys: add skill %d to user %d: %v", skillID, userID, err)
		}
	}
}

func (app *Service) deleteRecsysUserSkills(ctx context.Context, userID int, skillIDs []int) {
	for _, skillID := range skillIDs {
		if err := app.recsys.DeleteUserSkill(ctx, userID, skillID); err != nil {
			log.Printf("recsys: delete skill %d from user %d: %v", skillID, userID, err)
		}
	}
}

func (app *Service) addRecsysInternshipSkills(ctx context.Context, internshipID int, skillIDs []int) {
	for _, skillID := range skillIDs {
		if err := app.recsys.AddInternshipSkill(ctx, internshipID, skillID); err != nil {
			log.Printf("recsys: add skill %d to internship %d: %v", skillID, internshipID, err)
		}
	}
}

func (app *Service) deleteRecsysInternshipSkills(ctx context.Context, internshipID int, skillIDs []int) {
	for _, skillID := range skillIDs {
		if err := app.recsys.DeleteInternshipSkill(ctx, internshipID, skillID); err != nil {
			log.Printf("recsys: delete skill %d from internship %d: %v", skillID, internshipID, err)
		}
	}
}

// sendRecsysAction reports a user action (click / apply / invate) to recsys.
func (app *Service) sendRecsysAction(ctx context.Context, userID, internshipID int, actionType string) {
	err := app.recsys.AddAction(ctx, userID, internshipID, actionType)
	if err != nil {
		log.Printf("recsys: action %s user %d internship %d: %v", actionType, userID, internshipID, err)
	}
}

// GetRecommendations returns ranked internship recommendations for the current intern.
func (app *Service) GetRecommendations(ctx context.Context) ([]models.Recommendation, error) {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return nil, apierrors.ErrInternalServer
	}

	recommendations, err := app.recsys.GetRecommendations(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	return recommendations, nil
}

// TrackInternshipView reports a click event for the current intern.
func (app *Service) TrackInternshipView(ctx context.Context, internshipID int) {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return
	}
	app.sendRecsysAction(ctx, claims.UserID, internshipID, constants.RecsysActionClick)
}
