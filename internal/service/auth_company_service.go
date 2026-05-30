// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (app *Service) AuthCompany(ctx context.Context, userReg models.CompanyAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	exist, err := app.db.CheckUserEmail(ctx, userReg.Email, constants.Company)
	if err != nil {
		return "", err
	}
	if exist {
		return "", apierrors.ErrUserAlreadyExists
	}

	companyData, err := app.dadata.ValidTIN(userReg.Inn)
	if err != nil {
		return "", err
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrInternalServer, err)
	}

	sessionID := uuid.New().String()

	code, err := utils.NewCode()
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrInternalServer, err)
	}

	err = app.email.SendVerifyCode(userReg.Email, code)
	if err != nil {
		return "", err
	}

	user := models.CompanyVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
		DeviceID:     userReg.DeviceID,
		CompanyData:  companyData,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(ctx, sessionID, userJson, constants.VerifyCodeTimeLife)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrTooManyInvalidAttempts, err)
	}

	return sessionID, nil

}

func (app *Service) VerifyCompany(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	companyVerify := models.CompanyVerify{}
	res, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return "", "", apierrors.Warp(apierrors.ErrCodeExpired, err)
		}
		return "", "", err
	}

	err = json.Unmarshal([]byte(res), &companyVerify)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrJSONDecodeFailed, err)
	}

	if companyVerify.Code != verifyRequest.Code {
		if err != nil {
			return "", "", err
		}
		return "", "", apierrors.ErrInvalidCode
	}

	user := models.User{
		Email:        companyVerify.Email,
		Role:         constants.Company,
		PasswordHash: companyVerify.PasswordHash,
	}

	id, err := app.db.QueryNewCompany(ctx, user, companyVerify.CompanyData)

	user.Id = id
	if err != nil {
		return "", "", err
	}

	accessToken, err := utils.GeneratingJWT(id, companyVerify.DeviceID, user.Email, constants.Company, constants.AccessTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GeneratingJWT(id, companyVerify.DeviceID, user.Email, constants.Company, constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d.%s", id, companyVerify.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}
	app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, nil
}

func (app *Service) LoginCompany(ctx context.Context, userLog models.UserAuth) (string, string, error) {
	err := app.validate.Struct(userLog)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	ok, err := app.db.CheckUserEmail(ctx, userLog.Email, constants.Company)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", apierrors.ErrUserNotExists
	}

	user := models.User{}
	user, err = app.db.GetUser(ctx, userLog.Email, constants.Company)
	if err != nil {
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLog.Password))
	if err != nil {
		return "", "", apierrors.ErrInvalidPassword
	}

	accessToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, constants.Company, constants.AccessTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, constants.Company, constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d.%s", user.Id, userLog.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
