// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"

	"context"
)

func (app *Service) AuthCompany(ctx context.Context, userReg models.CompanyAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	companyData, err := app.dadata.ValidTIN(userReg.Inn)
	if err != nil {
		return "", err
	}

	userJSON, err := encodeVerifyJSON(userReg.Email, userReg.DeviceID, userReg.Password, nil, &companyData)

	sessionID, err := app.authUser(ctx, userJSON, userReg.Email, constants.Company)

	return sessionID, err
}
