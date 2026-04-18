package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db       *repository.Repository
	cache    *cache.Redis
	email    *EmailClient
	validate *validator.Validate
}

func NewAuthService(db *repository.Repository, cache *cache.Redis, email *EmailClient, validate *validator.Validate) *AuthService {
	return &AuthService{
		db:       db,
		cache:    cache,
		email:    email,
		validate: validate,
	}
}

func (app *AuthService) AuthUser(ctx context.Context, userReg models.UserAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	exist, err := app.db.CheckUserEmail(ctx, userReg.Email, "internal")
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	if exist {
		return "", fmt.Errorf("%w", apierrors.ErrUserAlreadyExists)
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

	user := models.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
		DeviceID:     userReg.DeviceID,
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

func (app *AuthService) VerifyUser(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	userVerify := models.UserVerify{}
	res, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCodeExpired, err)
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	if userVerify.CodeAmount >= 3 {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	if userVerify.Attempts >= 3 {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	if userVerify.Code != verifyRequest.Code {
		userVerify.Attempts++
		_, err := app.cache.Do(ctx, sessionID, "$.attempts", userVerify.Attempts)
		if err != nil {
			return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
		}
		return "", "", fmt.Errorf("%w", apierrors.ErrInvalidCode)
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         "internal",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := app.db.QueryNewUser(ctx, user)
	user.Id = id
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	accessToken, err := utils.GeneratingJWT(id, userVerify.DeviceID, user.Email, "internal", time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(id, userVerify.DeviceID, user.Email, "internal", time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d.%s", id, userVerify.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
	}
	app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, nil
}

func (app *AuthService) NewCode(ctx context.Context, sessionID string) error {
	newCode, err := utils.NewCode()
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	res, err := app.cache.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}
	user := models.UserVerify{}
	json.Unmarshal([]byte(res), &user)

	app.email.SendVerifyCode(user.Email, newCode)

	app.cache.Do(ctx, sessionID, "$.code", newCode)
	app.cache.Do(ctx, sessionID, "$.code_amount", user.CodeAmount+1)
	return nil
}

func (app *AuthService) LoginUser(ctx context.Context, userLog models.UserAuth) (string, string, error) {
	err := app.validate.Struct(userLog)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	ok, err := app.db.CheckUserEmail(ctx, userLog.Email, "internal")
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

	accessToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, "internal", time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(user.Id, userLog.DeviceID, user.Email, "internal", time.Hour*24*7)
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
