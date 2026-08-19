// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

// ErrorHandler is a handler function that can return an error.
type ErrorHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

// ErrorMiddleware wraps an ErrorHandler and converts returned errors into HTTP responses.
func ErrorMiddleware(next ErrorHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := next(w, r, ps); err != nil {
			status, message := apierrors.HTTPStatusMapping(err)
			http.Error(w, message, status)
			log.Println(err)
		}
	}
}

// CorsMiddleware sets CORS headers and handles preflight OPTIONS requests.
func (h *ServiceHandler) CorsMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		w.Header().Set("Access-Control-Allow-Origin", h.corsConfig.ControlAllowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.corsConfig.ControlAllowHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return next(w, r, ps)
	}
}

// AuthMiddleware authenticates the request, refreshes expired tokens, and checks for banned users.
func (h *ServiceHandler) AuthMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {

		tokenStr, err := h.getAuthTokenFromHeader(r)
		if err != nil {
			return err
		}

		token, claims, err := utils.DecodeJWT(tokenStr)
		if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
			return apierrors.Wrap(apierrors.ErrUnauthorized, err)
		}

		reason, err := h.service.HandlingBannedUser(r.Context(), claims.UserID)
		if err != nil {
			body, encodeErr := json.Marshal(models.BanResponse{Reason: reason})
			if encodeErr != nil {
				return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, encodeErr)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, writeErr := w.Write(body)
			if writeErr != nil {
				return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, writeErr)
			}
			return nil
		}

		ctx := context.WithValue(r.Context(), constants.ClaimsKey, claims)

		if !token.Valid {

			err = h.refreshTokenValid(ctx, r)
			if err != nil {
				return err
			}

			newToken, err := utils.GeneratingJWT(claims.UserID, claims.DeviceID, claims.Email, claims.Role, constants.AccessTokenTimeLife)
			if err != nil {
				return err
			}

			w.Header().Set("X-New-Access-Token", newToken)
			return next(w, r.WithContext(ctx), ps)

		}

		return next(w, r.WithContext(ctx), ps)
	}
}

func (h ServiceHandler) getAuthTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	return "", apierrors.ErrUnauthorized
}

func (h ServiceHandler) refreshTokenValid(ctx context.Context, r *http.Request) error {
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	refreshToken, _, err := utils.DecodeJWT(refreshTokenCookie.Value)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrUnauthorized, err)
	}

	refreshTokenCache, err := h.service.GetRefreshToken(ctx)
	if err != nil {
		return err
	}
	if refreshTokenCookie.Value != refreshTokenCache {
		return apierrors.ErrUnauthorized
	}

	if !refreshToken.Valid {
		return apierrors.ErrUnauthorized
	}
	return nil
}

// InternMiddleware restricts the request to authenticated interns.
func (h *ServiceHandler) InternMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		claims, ok := r.Context().Value(constants.ClaimsKey).(models.Claims)
		if !ok {
			return apierrors.ErrInternalServer
		}
		if claims.Role != constants.Intern {
			return apierrors.ErrForbidden
		}
		return next(w, r, ps)
	}
}

// AdminMiddleware restricts the request to authenticated admins.
func (h *ServiceHandler) AdminMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		claims, ok := r.Context().Value(constants.ClaimsKey).(models.Claims)
		if !ok {
			return apierrors.ErrInternalServer
		}
		if claims.Role != constants.Admin {
			return apierrors.ErrForbidden
		}
		return next(w, r, ps)
	}
}

// CompanyMiddleware restricts the request to authenticated companies.
func (h *ServiceHandler) CompanyMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		claims, ok := r.Context().Value(constants.ClaimsKey).(models.Claims)
		if !ok {
			return apierrors.ErrInternalServer
		}
		if claims.Role != constants.Company {
			return apierrors.ErrForbidden
		}
		return next(w, r, ps)
	}
}

// clientIP extracts the client IP from a RemoteAddr, handling IPv6 addresses
// correctly and normalizing IPv4-mapped IPv6 addresses to plain IPv4.
func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr.Unmap().String()
	}
	return host
}

// RateLimitMiddleware rate limits requests to an endpoint by user, session, or IP.
func (h *ServiceHandler) RateLimitMiddleware(next ErrorHandler, rate int, endpoint string) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		ctx := r.Context()
		var id string
		claimsContext := ctx.Value(constants.ClaimsKey)
		sessionID := r.Header.Get("X-Verify-Session")
		if claimsContext != nil {
			claims, ok := claimsContext.(models.Claims)
			if !ok {
				return apierrors.ErrInternalServer
			}
			id = strconv.Itoa(claims.UserID)
		} else if sessionID != "" {
			id = sessionID
		} else {
			id = clientIP(r.RemoteAddr)
		}
		key := id + "." + endpoint
		ok, err := h.service.RateLimitCheck(ctx, key, rate)
		if err != nil {
			return err
		}
		if !ok {
			return apierrors.ErrTooManyInvalidAttempts
		}
		return next(w, r, ps)
	}
}
