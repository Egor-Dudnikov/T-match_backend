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

func (app *AuthService) AuthUser(userReg models.UserRegistration) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrValidationFailed, err)
	}

	exist, err := app.db.CheckUserEmail(userReg.Email)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrDatabaseError, err)
	}
	if exist {
		return "", fmt.Errorf("%v", apierrors.ErrUserAlreadyExists)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrInternalServer, err)
	}

	sesionToken := uuid.New().String()

	code, err := utils.NewCode()
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrInternalServer, err)
	}

	err = app.email.SendVerifyCode(userReg.Email, code)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrEmailSendFailed, err)
	}

	user := models.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
		DeviceID:     userReg.DeviceID,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(sesionToken, userJson, time.Minute*7)
	if err != nil {
		return "", fmt.Errorf("%v: %v", apierrors.ErrCacheError, err)
	}

	return sesionToken, nil

}

func (app *AuthService) VerifyStudent(sesionToken string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrValidationFailed, err)
	}

	userVerify := models.UserVerify{}
	res, err := app.cache.Get(sesionToken)
	defer app.cache.Del(sesionToken)

	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrCodeExpired, err)
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	if userVerify.Code != verifyRequest.Code {
		return "", "", fmt.Errorf("%v", apierrors.ErrInvalidCode, err)
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         "intern",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := app.db.QueeryNewUser(user)
	user.Id = id
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrDatabaseError, err)
	}

	accessToken, err := utils.GeneratingJWT(strconv.Itoa(id), userVerify.DeviceID, user.Email, time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	refreshToken, err := utils.GeneratingJWT(strconv.Itoa(id), userVerify.DeviceID, user.Email, time.Hour*24*7)
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrJWTGenerationFailed, err)
	}

	key := fmt.Sprintf("%d%s", id, userVerify.DeviceID)
	err = app.cache.Set(key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("%v: %v", apierrors.ErrCacheError, err)
	}
	return accessToken, refreshToken, nil
}
