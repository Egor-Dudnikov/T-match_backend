package service

import (
	"T-match_backend/internal/cache"
	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
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
		return "", err
	}

	exist, err := app.db.CheckUserEmail(userReg.Email)
	if err != nil {
		return "", err
	}
	if exist {
		return "", fmt.Errorf("User with this email exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	sesionToken, err := utils.GeneratingJWT("0", userReg.DeviceID, userReg.Email, 3*time.Minute)
	if err != nil {
		return "", err
	}

	code, err := utils.NewCode()
	if err != nil {
		return "", err
	}

	err = app.email.SendVerifyCode(userReg.Email, code)
	if err != nil {
		return "", err
	}

	user := models.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		return "", err
	}

	err = app.cache.Set(sesionToken, userJson, time.Minute*7)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return sesionToken, nil

}

func (app *AuthService) VerifyStudent(sesionToken string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", err
	}

	userVerify := models.UserVerify{}
	res, err := app.cache.Get(sesionToken)
	if err != nil {
		return "", "", err
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		return "", "", err
	}

	if userVerify.Code != verifyRequest.Code {
		return "", "", fmt.Errorf("Not right verify code")
	}

	app.cache.Del(sesionToken)

	_, claim, err := utils.DecodeJWT(sesionToken)
	if err != nil {
		return "", "", err
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         "intern",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := app.db.QueeryNewUser(user)
	user.Id = id
	if err != nil {
		return "", "", err
	}

	accessToken, err := utils.GeneratingJWT(string(id), claim.DeviceID, user.Email, time.Minute*15)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GeneratingJWT(string(id), claim.DeviceID, user.Email, time.Hour*24*7)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d%s", id, claim.DeviceID)
	err = app.cache.Set(key, []byte(refreshToken), 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
