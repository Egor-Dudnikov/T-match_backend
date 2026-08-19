// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/cache"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/dadata"
	"T-match_backend/internal/models"
	"T-match_backend/internal/recsys"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/s3"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Service aggregates the repositories, cache, clients and hub used to
// implement the application's business logic.
type Service struct {
	db       *repository.Repository
	cache    *cache.Redis
	email    *EmailClient
	s3       *s3.Storage
	dadata   *dadata.Client
	recsys   *recsys.Client
	Hub      *Hub
	validate *validator.Validate
}

// Newservice creates a Service with the given dependencies.
func Newservice(db *repository.Repository, cache *cache.Redis, email *EmailClient, validate *validator.Validate, s3 *s3.Storage, dadataclient *dadata.Client, hub *Hub, recsysClient *recsys.Client) *Service {
	return &Service{
		db:       db,
		cache:    cache,
		email:    email,
		validate: validate,
		s3:       s3,
		Hub:      hub,
		dadata:   dadataclient,
		recsys:   recsysClient,
	}
}

// RegService initializes the repositories, cache and clients from the given
// config and returns a ready-to-use Service.
func RegService(config models.Config) (*Service, error) {
	db, err := repository.PingDatabase(config.DbConfig)
	if err != nil {
		return &Service{}, err
	}

	dbr, err := cache.PingRedis(config.RedisConfig)
	if err != nil {
		return &Service{}, err
	}
	s3Client, err := s3.LoadS3(config.S3Config)
	if err != nil {
		return &Service{}, err
	}

	repo := repository.NewRepository(db)
	redis := cache.NewRedis(dbr)
	email := NewEmailClient(config.EmailConfig)

	s3Storage, err := s3.NewS3(s3Client, config.S3Config)
	if err != nil {
		return &Service{}, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return &Service{}, err
	}

	dadataClient := dadata.NewClient()

	recsysClient := recsys.NewClient(config.RecsysConfig.URL)

	hub := newHub()

	app := Newservice(repo, redis, email, validate, s3Storage, dadataClient, hub, recsysClient)
	return app, err
}

// CloseDB closes the underlying database connection.
func (app *Service) CloseDB() error {
	return app.db.Close()
}

// CloseRedis closes the underlying Redis connection.
func (app *Service) CloseRedis() error {
	return app.cache.Close()
}

// UserVerify defines the operations required to persist and authenticate a
// user pending email verification.
type UserVerify interface {
	QueryNewUser(ctx context.Context, db *repository.Repository) (int, error)
	GeneratingTokenPair(id int) (string, string, error)
	GetUserKey(id int) string
}

// InternVerify holds the data of an intern pending email verification.
type InternVerify struct {
	Email        string
	PasswordHash string
	DeviceID     string
	BirthDate    time.Time
}

// QueryNewUser creates the intern in the database and returns the new user ID.
func (iv InternVerify) QueryNewUser(ctx context.Context, db *repository.Repository) (int, error) {
	id, err := db.QueryNewUser(ctx, models.InternVerify{
		Email:        iv.Email,
		PasswordHash: iv.PasswordHash,
		DeviceID:     iv.DeviceID,
		BirthDate:    iv.BirthDate,
	})
	return id, err
}

// GeneratingTokenPair generates an access and refresh token pair for the intern
// with the given ID.
func (iv InternVerify) GeneratingTokenPair(id int) (string, string, error) {
	accessToken, refreshToken, err := utils.GeneratingTokenPair(id, iv.DeviceID, iv.Email, constants.Intern)
	return accessToken, refreshToken, err
}

// GetUserKey returns the cache key under which the intern's refresh token is
// stored.
func (iv InternVerify) GetUserKey(id int) string {
	return fmt.Sprintf("%d.%s", id, iv.DeviceID)
}

// AuthIntern validates an intern's registration data, sends a verification code
// and returns the verification session ID.
func (app *Service) AuthIntern(ctx context.Context, userReg models.InternAuth) (string, error) {
	err := app.validate.Struct(userReg)
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	if !utils.ValidAge(userReg.BirthDate) {
		return "", apierrors.ErrUserMustBe16
	}

	userJSON, err := encodeVerifyJSON(userReg.Email, userReg.DeviceID, userReg.Password, &userReg.BirthDate, nil)
	if err != nil {
		return "", err
	}

	sessionID, err := app.authUser(ctx, userJSON, userReg.Email, constants.Intern)

	return sessionID, err
}

func encodeVerifyJSON(email, deviceID, password string, birthDate *time.Time, companyData *models.CompanyData) ([]byte, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	var user interface{}

	if birthDate != nil {
		user = InternVerify{
			Email:        email,
			PasswordHash: string(hashPassword),
			DeviceID:     deviceID,
			BirthDate:    *birthDate,
		}
	} else if companyData != nil {
		user = CompanyVerify{
			Email:        email,
			PasswordHash: string(hashPassword),
			DeviceID:     deviceID,
			CompanyData:  *companyData,
		}
	} else {
		return []byte{}, apierrors.ErrInternalServer
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return userJSON, apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	return userJSON, nil
}

func (app *Service) authUser(ctx context.Context, userJSON []byte, email string, role string) (string, error) {

	exist, err := app.db.CheckUserEmail(ctx, email, role)
	if err != nil {
		return "", err
	}
	if exist {
		return "", apierrors.ErrUserAlreadyExists
	}

	sessionID := uuid.New().String()

	code, err := utils.NewCode()
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	err = app.email.SendVerifyCode(email, code)
	if err != nil {
		return "", err
	}

	err = app.cache.Set(ctx, sessionID, userJSON, constants.VerifyCodeTimeLife)
	if err != nil {
		return "", err
	}
	err = app.cache.Set(ctx, sessionID+".code", []byte(code), constants.VerifyCodeTimeLife)
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

/*
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
*/

func unmarshalUserVerify(userJSON string, role string) (UserVerify, error) {

	var err error
	switch role {
	case constants.Intern:
		var userVerify = InternVerify{}
		err = json.Unmarshal([]byte(userJSON), &userVerify)
		return userVerify, err
	case constants.Company:
		var userVerify = CompanyVerify{}
		err = json.Unmarshal([]byte(userJSON), &userVerify)
		return userVerify, err
	default:
		err = apierrors.ErrInternalServer
	}
	return nil, err
}

// VerifyUser checks the verification code, creates the user and returns access
// and refresh tokens.
func (app *Service) VerifyUser(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest, role string) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	userJSON, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return "", "", apierrors.Wrap(apierrors.ErrCodeExpired, err)
		}
		return "", "", err
	}

	userVerify, err := unmarshalUserVerify(userJSON, role)
	if err != nil {
		return "", "", err
	}

	err = app.verify(ctx, sessionID+".code", verifyRequest.Code)
	if err != nil {
		return "", "", err
	}

	id, err := userVerify.QueryNewUser(ctx, app.db)
	if err != nil {
		return "", "", err
	}

	if role == constants.Intern {
		app.createRecsysUser(ctx, id)
	}

	accessToken, refreshToken, err := userVerify.GeneratingTokenPair(id)
	if err != nil {
		return "", "", err
	}

	key := userVerify.GetUserKey(id)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}
	err = app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, err
}

// NewCode generates a new verification code for the given session and resends
// it to the user's email.
func (app *Service) NewCode(ctx context.Context, sessionID string) error {
	newCode, err := utils.NewCode()
	if err != nil {
		return apierrors.Wrap(apierrors.ErrInternalServer, err)
	}

	res, err := app.cache.Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return apierrors.Wrap(apierrors.ErrSessionExpired, err)
		}
		return err
	}

	user := InternVerify{}
	err = json.Unmarshal([]byte(res), &user)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
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

// LoginUser validates the user's credentials and returns access and refresh
// tokens.
func (app *Service) LoginUser(ctx context.Context, userLog models.LoginUser, role string) (string, string, error) {
	err := app.validate.Struct(userLog)
	if err != nil {
		return "", "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	ok, err := app.db.CheckUserEmail(ctx, userLog.Email, role)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", apierrors.ErrUserNotExists
	}

	user, err := app.db.GetUser(ctx, userLog.Email, role)
	if err != nil {
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLog.PasswordHash))
	if err != nil {
		return "", "", apierrors.ErrInvalidPassword
	}

	accessToken, refreshToken, err := utils.GeneratingTokenPair(user.ID, userLog.DeviceID, user.Email, role)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d.%s", user.ID, userLog.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// FogotPassword sends a verification code to the user's email for password
// reset and returns the verification session ID.
func (app *Service) FogotPassword(ctx context.Context, user models.FogetPasswordRequest) (string, error) {
	err := app.validate.Struct(user)
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	if exists, err := app.db.CheckUserEmail(ctx, user.Email, user.Role); err == nil && !exists {
		return "", apierrors.ErrUserNotExists
	} else if err != nil {
		return "", err
	}

	id, err := app.db.GetUserIDByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	code, err := utils.NewCode()
	if err != nil {
		return "", nil
	}

	sessionID := uuid.New().String()

	err = app.cache.Set(ctx, sessionID+".code", []byte(code), constants.VerifyCodeTimeLife)
	if err != nil {
		return "", err
	}

	userInfo := models.UserInfo{
		UserID:   id,
		Email:    user.Email,
		Role:     user.Role,
		DeviceID: user.DeviceID,
	}

	userJSON, err := json.Marshal(userInfo)
	if err != nil {
		return "", apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}

	err = app.cache.Set(ctx, sessionID, userJSON, constants.VerifyCodeTimeLife)
	if err != nil {
		return "", err
	}

	err = app.email.SendVerifyCode(user.Email, code)

	return sessionID, err
}

// VerifyFogottenUser verifies the reset code and returns access and refresh
// tokens for the user.
func (app *Service) VerifyFogottenUser(ctx context.Context, sessionID string, verifyRequest models.VerifyRequest) (string, string, error) {
	err := app.validate.Struct(verifyRequest)
	if err != nil {
		return "", "", apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	userVerify := InternVerify{}
	res, err := app.cache.Get(ctx, sessionID)

	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return "", "", apierrors.Wrap(apierrors.ErrCodeExpired, err)
		}
		return "", "", err
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		return "", "", apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}

	err = app.verify(ctx, sessionID+".code", verifyRequest.Code)
	if err != nil {
		return "", "", err
	}

	userJSON, err := app.cache.Get(ctx, sessionID)
	if err != nil {
		return "", "", err
	}

	userInfo := models.UserInfo{}

	err = json.Unmarshal([]byte(userJSON), &userInfo)
	if err != nil {
		return "", "", apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}

	accessToken, refreshToken, err := utils.GeneratingTokenPair(userInfo.UserID, userInfo.DeviceID, userInfo.Email, userInfo.Role)
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%d.%s", userInfo.UserID, userVerify.DeviceID)
	err = app.cache.Set(ctx, key, []byte(refreshToken), constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}
	err = app.cache.Del(ctx, sessionID)
	return accessToken, refreshToken, err
}

// ChangePassword updates the password of the authenticated user.
func (app *Service) ChangePassword(ctx context.Context, newPasswordReq models.ChangePasswordRequest) error {
	err := app.validate.Struct(newPasswordReq)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	newPassword := newPasswordReq.Password

	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = app.db.UpdatePasswordHash(ctx, string(passwordHash), claims.UserID)
	return err
}

// GetRefreshToken returns the refresh token of the authenticated user.
func (app *Service) GetRefreshToken(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return "", apierrors.ErrInternalServer
	}

	key := fmt.Sprintf("%d.%s", claims.UserID, claims.DeviceID)
	token, err := app.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, apierrors.ErrKeyNotFound) {
			return token, apierrors.Wrap(apierrors.ErrSessionExpired, err)
		}
		return token, err
	}

	return token, nil
}

// RateLimitCheck reports whether the given key has exceeded the rate limit.
func (app *Service) RateLimitCheck(ctx context.Context, key string, rate int) (bool, error) {
	ok, err := app.cache.RateLimitCheck(ctx, key, rate)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// DeleteRefreshToken deletes the refresh token of the authenticated user.
func (app *Service) DeleteRefreshToken(ctx context.Context) error {
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	err := app.cache.Del(ctx, fmt.Sprintf("%d.%s", claims.UserID, claims.DeviceID))
	return err
}
