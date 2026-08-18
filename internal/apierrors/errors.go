// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package apierrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotExists     = errors.New("user not exists")

	ErrInvalidCode    = errors.New("invalid verification code")
	ErrCodeExpired    = errors.New("verification code expired")
	ErrSessionExpired = errors.New("verification session expired")

	ErrDatabaseError = errors.New("database error")

	ErrCacheError = errors.New("cache error")

	ErrEmailSendFailed = errors.New("failed to send email")

	ErrJWTGenerationFailed = errors.New("failed to generate JWT")
	ErrJWTDecodingFailed   = errors.New("failed to decode JWT")

	ErrJSONDecodeFailed = errors.New("failed to decode JSON")
	ErrJSONEncodeFailed = errors.New("failed to encode JSON")

	ErrBadRequest             = errors.New("bad request")
	ErrInternalServer         = errors.New("internal server error")
	ErrTooManyInvalidAttempts = errors.New("too many invalid attempts")

	ErrInvalidPassword = errors.New("invalid password")

	ErrBadGateway       = errors.New("bad gateway")
	ErrCompanyNotExists = errors.New("company not exists")
	ErrUnauthorized     = errors.New("user unauthorized")

	ErrForbidden    = errors.New("access forbidden")
	ErrUserMustBe16 = errors.New("birth date is invalid")

	ErrInternshipIsArchived = errors.New("internship is archived")
	ErrInternshipNotFound   = errors.New("internship not found")
	ErrAlreadyResponded     = errors.New("already responded to this internship")
	ErrCityNotFound         = errors.New("city not found")

	ErrSkillsNotFound     = errors.New("skills not found")
	ErrKeyNotFound        = errors.New("redis key not found")
	ErrCompanyNameMissing = errors.New("company have not name")

	ErrUserAlreadyBanned = errors.New("user already baned")
	ErrUserNotBanned     = errors.New("user not baned")
	ErrUserBanned        = errors.New("user baned")
	ErrUserNotFound      = errors.New("user not found")
	ErrCannotBanAdmin    = errors.New("cannot ban admin")
	ErrProfileNotFound   = errors.New("profile not found")
	ErrCompanyNotFound   = errors.New("company not found")
)

type ErrorMapping struct {
	Status  int
	Message string
}

var errorStatusMap = map[error]ErrorMapping{
	ErrInvalidCode:        {http.StatusBadRequest, "Invalid verification code format"},
	ErrCodeExpired:        {http.StatusBadRequest, "Verification code expired"},
	ErrSessionExpired:     {http.StatusBadRequest, "Verification session expired"},
	ErrJSONDecodeFailed:   {http.StatusBadRequest, "Invalid JSON format"},
	ErrJSONEncodeFailed:   {http.StatusBadRequest, "Failed to process data"},
	ErrBadRequest:         {http.StatusBadRequest, "Bad request"},
	ErrCompanyNameMissing: {http.StatusBadRequest, "company have not name"},

	ErrInvalidPassword: {http.StatusUnauthorized, "Invalid password"},
	ErrUnauthorized:    {http.StatusUnauthorized, "User Unauthorized"},

	ErrForbidden: {http.StatusForbidden, "Access denied: insufficient permissions"},

	ErrUserAlreadyExists: {http.StatusConflict, "User with this email already exists"},
	ErrAlreadyResponded:  {http.StatusConflict, "You have already responded to this internship"},

	ErrUserNotExists:      {http.StatusNotFound, "User with this email not exists"},
	ErrCompanyNotExists:   {http.StatusNotFound, "Company with this TIN not exists"},
	ErrInternshipNotFound: {http.StatusNotFound, "Internship not found"},
	ErrSkillsNotFound:     {http.StatusNotFound, "Skills not found"},
	ErrCityNotFound:       {http.StatusNotFound, "City not found"},

	ErrInternshipIsArchived: {http.StatusGone, "Internship is archived"},

	ErrUserMustBe16: {http.StatusUnprocessableEntity, "User must be at least 16 years old"},

	ErrTooManyInvalidAttempts: {http.StatusTooManyRequests, "Too many invalid attempts"},

	ErrCacheError:      {http.StatusServiceUnavailable, "Cache service temporarily unavailable"},
	ErrEmailSendFailed: {http.StatusServiceUnavailable, "Failed to send email, please try again"},

	ErrBadGateway: {http.StatusBadGateway, "External service temporarily unavailable. Please try again later."},

	ErrDatabaseError:       {http.StatusInternalServerError, "Internal server error"},
	ErrJWTGenerationFailed: {http.StatusInternalServerError, "Internal server error"},
	ErrJWTDecodingFailed:   {http.StatusInternalServerError, "Internal server error"},
	ErrInternalServer:      {http.StatusInternalServerError, "Internal server error"},

	ErrUserAlreadyBanned: {http.StatusConflict, "User is already banned"},
	ErrUserBanned:        {http.StatusConflict, "User is banned"},
	ErrUserNotBanned:     {http.StatusConflict, "User not baned"},
	ErrUserNotFound:      {http.StatusNotFound, "User not found"},
	ErrCannotBanAdmin:    {http.StatusBadRequest, "Admins cannot be banned"},
	ErrProfileNotFound:   {http.StatusNotFound, "Profile not found"},
	ErrCompanyNotFound:   {http.StatusNotFound, "Company not found"},
}

func HTTPStatusMapping(err error) (status int, message string) {
	for e, m := range errorStatusMap {
		if errors.Is(err, e) {
			return m.Status, m.Message
		}
	}
	deafult := errorStatusMap[ErrInternalServer]
	return deafult.Status, deafult.Message
}

func Wrap(apierror error, err error) error {
	return fmt.Errorf("%w: %w", apierror, err)
}
