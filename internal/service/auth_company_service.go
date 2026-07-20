// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/utils"
	"encoding/json"
	"fmt"
	"time"

	"context"
)

type CompanyVerify struct {
	CompanyData  models.CompanyData
	Email        string
	PasswordHash string
	DeviceID     string
}

func GetCompanyVerify(ctx context.Context, userJSON string) (UserVerify, error) {
	userVerify := CompanyVerify{}
	err := json.Unmarshal([]byte(userJSON), &userVerify)
	return userVerify, err
}

func (cv CompanyVerify) QueryNewUser(ctx context.Context, db *repository.Repository) (int, error) {
	id, err := db.QueryNewCompany(ctx, models.CompanyVerify{
		Email:        cv.Email,
		PasswordHash: cv.PasswordHash,
		DeviceID:     cv.DeviceID,
		CompanyData:  cv.CompanyData,
	})
	return id, err
}

func (cv CompanyVerify) GeneratingJWT(id int, timeLife time.Duration) (string, error) {
	token, err := utils.GeneratingJWT(id, cv.DeviceID, cv.Email, constants.Company, timeLife)
	return token, err
}

func (cv CompanyVerify) GetUserKey(id int) string {
	return fmt.Sprintf("%d.%s", id, cv.DeviceID)
}

func (app *Service) AuthCompany(ctx context.Context, userReg models.CompanyAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	companyData, err := app.dadata.ValidTIN(userReg.Inn)
	if err != nil {
		return "", err
	}

	userJSON, err := encodeVerifyJSON(userReg.Email, userReg.DeviceID, userReg.Password, nil, &companyData)
	if err != nil {
		return "", err
	}

	sessionID, err := app.authUser(ctx, userJSON, userReg.Email, constants.Company)

	return sessionID, err
}
