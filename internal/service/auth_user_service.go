// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/dadata"
	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/s3"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db       *repository.Repository
	cache    *cache.Redis
	email    *EmailClient
	s3       *s3.S3Storage
	dadata   *dadata.DadataClient
	validate *validator.Validate
}

func Newservice(db *repository.Repository, cache *cache.Redis, email *EmailClient, validate *validator.Validate, s3 *s3.S3Storage, dadataclient *dadata.DadataClient) *Service {
	return &Service{
		db:       db,
		cache:    cache,
		email:    email,
		validate: validate,
		s3:       s3,
		dadata:   dadataclient,
	}
}

func (app *Service) AuthUser(ctx context.Context, userReg models.UserAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	if !utils.ValidAge(userReg.BirthDate) {
		return "", apierrors.ErrUserMustBe16
	}

	exist, err := app.db.CheckUserEmail(ctx, userReg.Email, constants.Intern)
	if err != nil {
		return "", err
	}
	if exist {
		return "", apierrors.ErrUserAlreadyExists
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

	user := models.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		DeviceID:     userReg.DeviceID,
		BirthDate:    userReg.BirthDate,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", apierrors.Warp(apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(ctx, sessionID, userJson, constants.VerifyCodeTimeLife)
	if err != nil {
		return "", err
	}
	err = app.cache.Set(ctx, sessionID+".code", userJson, constants.VerifyCodeTimeLife)
	if err != nil {
		return "", err
	}

	return sessionID, nil

}

func (app *Service) verify(ctx context.Context, key string, code string) error {
	codeRedis, err := app.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return apierrors.ErrCodeExpired
		}
		return err
	}
	if codeRedis != code {
		return apierrors.ErrInvalidCode
	}
	return nil
}

func (app *Service) resetCode(ctx context.Context, key, newCode string) error {
	err := app.cache.ResetCode(ctx, key, newCode)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return apierrors.ErrSessionExpired
		}
		return err
	}
	return nil
}

func (app *Service) VerifyUser(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	userVerify := models.UserVerify{}
	res, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return "", "", apierrors.Warp(apierrors.ErrCodeExpired, err)
		}
		return "", "", err
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrJSONDecodeFailed, err)
	}

	err = app.verify(ctx, sessionID+".code", verifyRequest.Code)
	if err != nil {
		return "", "", err
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         constants.Intern,
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := app.db.QueryNewUser(ctx, user, userVerify.BirthDate)
	user.Id = id
	if err != nil {
		return "", "", err
	}

	accessToken, err := utils.GeneratingJWT(id, userVerify.DeviceID, user.Email, constants.Intern, constants.AccessTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GeneratingJWT(id, userVerify.DeviceID, user.Email, constants.Intern, constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d.%s", id, userVerify.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}
	app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, nil
}

func (app *Service) NewCode(ctx context.Context, sessionID string) error {
	newCode, err := utils.NewCode()
	if err != nil {
		return apierrors.Warp(apierrors.ErrInternalServer, err)
	}

	res, err := app.cache.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return apierrors.Warp(apierrors.ErrSessionExpired, err)
		}
		return err
	}

	user := models.UserVerify{}
	err = json.Unmarshal([]byte(res), &user)
	if err != nil {
		return apierrors.Warp(apierrors.ErrJSONDecodeFailed, err)
	}

	err = app.cache.ResetCode(ctx, sessionID+".code", newCode)
	if err != nil {
		return err
	}

	err = app.email.SendVerifyCode(user.Email, newCode)
	if err != nil {
		return err
	}

	return nil
}

func (app *Service) LoginUser(ctx context.Context, userLog models.UserAuth) (string, string, error) {
	err := app.validate.Struct(userLog)
	if err != nil {
		return "", "", apierrors.Warp(apierrors.ErrBadRequest, err)
	}

	ok, err := app.db.CheckUserEmail(ctx, userLog.Email, constants.Intern)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", apierrors.ErrUserNotExists
	}

	user := models.User{}
	user, err = app.db.GetUser(ctx, userLog.Email, constants.Intern)
	if err != nil {
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLog.Password))
	if err != nil {
		return "", "", apierrors.ErrInvalidPassword
	}

	accessToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, constants.Intern, constants.AccessTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, constants.Intern, constants.RefreshTokenTimeLife)
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

func (app *Service) GetRefreshToken(ctx context.Context, id int, deviceID string) (string, error) {
	key := fmt.Sprintf("%d.%s", id, deviceID)
	token, err := app.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return token, apierrors.Warp(apierrors.ErrSessionExpired, err)
		}
		return token, err
	}
	return token, nil
}

func (app *Service) RateLimitCheck(ctx context.Context, key string, rate int) (bool, error) {
	ok, err := app.cache.RateLimitCheck(ctx, key, rate)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (app *Service) DeleteRefreshToken(ctx context.Context) {
	claims := ctx.Value("claims").(models.Claims)
	app.cache.Del(ctx, claims.ID)
}
