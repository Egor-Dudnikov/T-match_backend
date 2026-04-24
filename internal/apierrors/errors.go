package apierrors

import (
	"errors"
	"net/http"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotExists     = errors.New("user not exists")

	ErrInvalidCode = errors.New("invalid verification code")
	ErrCodeExpired = errors.New("verification code expired")

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

	ErrForbidden = errors.New("access forbidden")
)

func HTTPStatusMapping(err error) (status int, message string) {
	switch {
	case errors.Is(err, ErrInvalidCode):
		return http.StatusBadRequest, "Invalid verification code format"
	case errors.Is(err, ErrCodeExpired):
		return http.StatusBadRequest, "Verification code expired"
	case errors.Is(err, ErrJSONDecodeFailed):
		return http.StatusBadRequest, "Invalid JSON format"
	case errors.Is(err, ErrJSONEncodeFailed):
		return http.StatusBadRequest, "Failed to process data"
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, "Bad request"

	// 401 Unauthorized
	case errors.Is(err, ErrInvalidPassword):
		return http.StatusUnauthorized, "Invalid password"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "User Unauthorized"

	// 403 Forbidden
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "Access denied: insufficient permissions"

	// 409 Conflict
	case errors.Is(err, ErrUserAlreadyExists):
		return http.StatusConflict, "User with this email already exists"

	// 404 Not Found
	case errors.Is(err, ErrUserNotExists):
		return http.StatusNotFound, "User with this email not exists"
	case errors.Is(err, ErrCompanyNotExists):
		return http.StatusNotFound, "Company with this TIN not exists"

	// 429 Too Many Request
	case errors.Is(err, ErrTooManyInvalidAttempts):
		return http.StatusTooManyRequests, "Too many invalid attempts"
	// 503 Service Unavailable
	case errors.Is(err, ErrCacheError):
		return http.StatusServiceUnavailable, "Cache service temporarily unavailable"
	case errors.Is(err, ErrEmailSendFailed):
		return http.StatusServiceUnavailable, "Failed to send email, please try again"

	// 502 Bad Gateway
	case errors.Is(err, ErrBadGateway):
		return http.StatusBadGateway, "External service temporarily unavailable. Please try again later."

	// 500 Internal Server Error
	case errors.Is(err, ErrDatabaseError):
		return http.StatusInternalServerError, "Internal server error"
	case errors.Is(err, ErrJWTGenerationFailed):
		return http.StatusInternalServerError, "Internal server error"
	case errors.Is(err, ErrJWTDecodingFailed):
		return http.StatusInternalServerError, "Internal server error"
	case errors.Is(err, ErrInternalServer):
		return http.StatusInternalServerError, "Internal server error"

	// Fallback
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
