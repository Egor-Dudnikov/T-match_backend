package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/utils"
	"encoding/json"
	"fmt"
	"strconv"
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

func (app *AuthService) AuthUser(userReg models.UserAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrValidationFailed, err)
	}

	exist, err := app.db.CheckUserEmail(userReg.Email)
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
		Attemts:      0,
		CodeAmount:   0,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(sessionID, userJson, time.Minute*7)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	return sessionID, nil

}

func (app *AuthService) VerifyStudent(sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrValidationFailed, err)
	}

	userVerify := models.UserVerify{}
	res, err := app.cache.Get(sessionID)

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

	if userVerify.Attemts >= 3 {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrTooManyInvalidAttempts, err)
	}

	if userVerify.Code != verifyRequest.Code {
		userVerify.Attemts++
		_, err := app.cache.Do(sessionID, "$.attemts", userVerify.Attemts)
		if err != nil {
			return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
		}
		if userVerify.Attemts >= 3 {
			return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
		}
		return "", "", fmt.Errorf("%w", apierrors.ErrInvalidCode, err)
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         "intern",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := app.db.QueryNewUser(user)
	user.Id = id
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	accessToken, err := utils.GeneratingJWT(strconv.Itoa(id), userVerify.DeviceID, user.Email, time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(strconv.Itoa(id), userVerify.DeviceID, user.Email, time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d%s", id, userVerify.DeviceID)
	err = app.cache.Set(key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
	}
	app.cache.Del(sessionID)
	return accessToken, refreshToken, nil
}

func (app *AuthService) NewCode(sessionID string) error {
	newCode, err := utils.NewCode()
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	res, err := app.cache.Get(sessionID)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}
	user := models.UserVerify{}
	json.Unmarshal([]byte(res), &user)

	app.email.SendVerifyCode(user.Email, newCode)

	app.cache.Do(sessionID, "$.code", newCode)
	app.cache.Do(sessionID, "$.code_amount", user.CodeAmount+1)
	return nil
}

func (app *AuthService) LoginStudent(userLog models.UserAuth) (string, string, error) {
	ok, err := app.db.CheckUserEmail(userLog.Email)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}
	if ok {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrUserNotExists)
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userLog.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}

	user := models.User{}
	user, err = app.db.GetUser(userLog.Email)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrDatabaseError, err)
	}

	if user.PasswordHash != string(hashPassword) {
		return "", "", fmt.Errorf("%w", apierrors.ErrInvalidPassword)
	}

	accessToken, err := utils.GeneratingJWT(strconv.Itoa(user.Id), userLog.DeviceID, user.Email, time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(strconv.Itoa(user.Id), userLog.DeviceID, user.Email, time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d%s", user.Id, userLog.DeviceID)
	err = app.cache.Set(key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", apierrors.ErrCacheError, err)
	}

	return accessToken, refreshToken, nil
}
