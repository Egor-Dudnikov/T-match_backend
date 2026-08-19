// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

// Package apierrors defines the error sentinels and HTTP status mappings used
// across the API layer.
package apierrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrUserAlreadyExists is returned when a user with the given email already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserNotExists is returned when no user matches the given email.
	ErrUserNotExists = errors.New("user not exists")

	// ErrInvalidCode is returned when the verification code format is invalid.
	ErrInvalidCode = errors.New("invalid verification code")
	// ErrCodeExpired is returned when the verification code has expired.
	ErrCodeExpired = errors.New("verification code expired")
	// ErrSessionExpired is returned when the verification session has expired.
	ErrSessionExpired = errors.New("verification session expired")

	// ErrDatabaseError is returned when a database operation fails.
	ErrDatabaseError = errors.New("database error")

	// ErrCacheError is returned when the cache service fails.
	ErrCacheError = errors.New("cache error")

	// ErrEmailSendFailed is returned when an email cannot be sent.
	ErrEmailSendFailed = errors.New("failed to send email")

	// ErrJWTGenerationFailed is returned when a JWT cannot be generated.
	ErrJWTGenerationFailed = errors.New("failed to generate JWT")
	// ErrJWTDecodingFailed is returned when a JWT cannot be decoded.
	ErrJWTDecodingFailed = errors.New("failed to decode JWT")

	// ErrJSONDecodeFailed is returned when a request body cannot be decoded.
	ErrJSONDecodeFailed = errors.New("failed to decode JSON")
	// ErrJSONEncodeFailed is returned when a response cannot be encoded.
	ErrJSONEncodeFailed = errors.New("failed to encode JSON")

	// ErrBadRequest is returned for malformed requests.
	ErrBadRequest = errors.New("bad request")
	// ErrInternalServer is returned for unexpected internal failures.
	ErrInternalServer = errors.New("internal server error")
	// ErrTooManyInvalidAttempts is returned after repeated invalid attempts.
	ErrTooManyInvalidAttempts = errors.New("too many invalid attempts")

	// ErrInvalidPassword is returned when the provided password is wrong.
	ErrInvalidPassword = errors.New("invalid password")

	// ErrBadGateway is returned when an upstream service fails.
	ErrBadGateway = errors.New("bad gateway")
	// ErrCompanyNotExists is returned when the company is not found.
	ErrCompanyNotExists = errors.New("company not exists")
	// ErrUnauthorized is returned for unauthenticated access.
	ErrUnauthorized = errors.New("user unauthorized")

	// ErrForbidden is returned when access is denied by role.
	ErrForbidden = errors.New("access forbidden")
	// ErrUserMustBe16 is returned when the user is younger than the allowed age.
	ErrUserMustBe16 = errors.New("birth date is invalid")

	// ErrInternshipIsArchived is returned when operating on an archived internship.
	ErrInternshipIsArchived = errors.New("internship is archived")
	// ErrInternshipNotFound is returned when an internship does not exist.
	ErrInternshipNotFound = errors.New("internship not found")
	// ErrAlreadyResponded is returned when a user responds twice to the same internship.
	ErrAlreadyResponded = errors.New("already responded to this internship")
	// ErrCityNotFound is returned when a city does not exist.
	ErrCityNotFound = errors.New("city not found")

	// ErrSkillsNotFound is returned when a skill is not found.
	ErrSkillsNotFound = errors.New("skills not found")
	// ErrKeyNotFound is returned when a redis key does not exist.
	ErrKeyNotFound = errors.New("redis key not found")
	// ErrCompanyNameMissing is returned when a company has no name.
	ErrCompanyNameMissing = errors.New("company have not name")

	// ErrUserAlreadyBanned is returned when banning an already banned user.
	ErrUserAlreadyBanned = errors.New("user already baned")
	// ErrUserNotBanned is returned when unbanning a user that is not banned.
	ErrUserNotBanned = errors.New("user not baned")
	// ErrUserBanned is returned when a banned user attempts an action.
	ErrUserBanned = errors.New("user baned")
	// ErrUserNotFound is returned when a user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrCannotBanAdmin is returned when attempting to ban an admin.
	ErrCannotBanAdmin = errors.New("cannot ban admin")
	// ErrProfileNotFound is returned when a profile does not exist.
	ErrProfileNotFound = errors.New("profile not found")
	// ErrCompanyNotFound is returned when a company does not exist.
	ErrCompanyNotFound = errors.New("company not found")
)

// ErrorMapping maps an error to its HTTP status and client-facing message.
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

// HTTPStatusMapping resolves an error into its HTTP status code and message.
func HTTPStatusMapping(err error) (status int, message string) {
	for e, m := range errorStatusMap {
		if errors.Is(err, e) {
			return m.Status, m.Message
		}
	}
	deafult := errorStatusMap[ErrInternalServer]
	return deafult.Status, deafult.Message
}

// Wrap wraps the apierror sentinel around an underlying error preserving both.
func Wrap(apierror error, err error) error {
	return fmt.Errorf("%w: %w", apierror, err)
}
