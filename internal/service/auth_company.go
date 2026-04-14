package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (app *AuthService) AuthCompany(ctx context.Context, userReg models.CompanyAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	exist, err := app.db.CheckUserEmail(ctx, userReg.Email, "company")
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	if exist {
		return "", fmt.Errorf("%w", apierrors.ErrUserAlreadyExists)
	}

	companyData, err := utils.ValidTIN(userReg.Inn)
	if err != nil {
		return "", err
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	sessionID := uuid.New().String()

	code, err := utils.NewCode()
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	err = app.email.SendVerifyCode(userReg.Email, code)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrEmailSendFailed, err)
	}

	user := models.CompanyVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
		DeviceID:     userReg.DeviceID,
		CompanyData:  companyData,
		Attempts:     0,
		CodeAmount:   0,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(ctx, sessionID, userJson, time.Minute*7)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	return sessionID, nil

}

func (app *AuthService) VerifyCompany(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	companyVerify := models.CompanyVerify{}
	res, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCodeExpired, err)
	}

	err = json.Unmarshal([]byte(res), &companyVerify)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	if companyVerify.CodeAmount >= 3 {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	if companyVerify.Attempts >= 3 {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	if companyVerify.Code != verifyRequest.Code {
		companyVerify.Attempts++
		_, err := app.cache.Do(ctx, sessionID, "$.attempts", companyVerify.Attempts)
		if err != nil {
			return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
		}
		return "", "", fmt.Errorf("%w", apierrors.ErrInvalidCode)
	}

	user := models.User{
		Email:        companyVerify.Email,
		Role:         "company",
		PasswordHash: companyVerify.PasswordHash,
	}

	id, err := app.db.QueryNewCompany(ctx, user, companyVerify.CompanyData)

	user.Id = id
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	accessToken, err := utils.GeneratingJWT(strconv.Itoa(id), companyVerify.DeviceID, user.Email, "company", time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(strconv.Itoa(id), companyVerify.DeviceID, user.Email, "company", time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d.%s", id, companyVerify.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
	}
	app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, nil
}

func (app *AuthService) LoginCompany(ctx context.Context, userLog models.UserAuth) (string, string, error) {
	err := app.validate.Struct(userLog)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	ok, err := app.db.CheckUserEmail(ctx, userLog.Email, "company")
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	if !ok {
		return "", "", fmt.Errorf("%w", apierrors.ErrUserNotExists)
	}

	user := models.User{}
	user, err = app.db.GetUser(ctx, userLog.Email)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLog.Password))
	if err != nil {
		return "", "", fmt.Errorf("%w", apierrors.ErrInvalidPassword)
	}

	accessToken, err := utils.GeneratingJWT(strconv.Itoa(user.Id), userLog.DeviceID, user.Email, "company", time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(strconv.Itoa(user.Id), userLog.DeviceID, user.Email, "company", time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d.%s", user.Id, userLog.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
	}

	return accessToken, refreshToken, nil
}
