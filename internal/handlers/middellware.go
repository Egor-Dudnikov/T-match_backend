package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"T-match_backend/internal/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func ErrorMiddleware(next ErrorHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := next(w, r, ps); err != nil {
			status, message := apierrors.HTTPStatusMapping(err)
			http.Error(w, message, status)
			log.Println(err)
		}
	}
}

func (h *AuthServiceHandler) CorsMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(h.corsConfig.ControlAllowOrigin, ", "))
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.corsConfig.ControlAllowHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return next(w, r, ps)
	}
}

func (h *AuthServiceHandler) AuthMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		tokenStr := r.Header.Get("Token")
		token, claims, err := utils.DecodeJWT(tokenStr)
		ctx := context.WithValue(r.Context(), "claims", claims)
		if err != nil {
			return fmt.Errorf("%w: %v", apierrors.ErrJWTDecodingFailed, err)
		}
		if !token.Valid {
			refreshTokenCookie, err := r.Cookie("refresh_token")
			if err != nil {
				return fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
			}
			refreshToken, _, err := utils.DecodeJWT(refreshTokenCookie.Value)
			if err != nil {
				return fmt.Errorf("%w: %v", apierrors.ErrJWTDecodingFailed, err)
			}
			if refreshToken.Valid {
				newToken, err := utils.GeneratingJWT(claims.UserID, claims.DeviceID, claims.Email, claims.Role, 15*time.Minute)
				if err != nil {
					return fmt.Errorf("%w: %v", apierrors.ErrJWTGenerationFailed, err)
				}
				w.Header().Set("Token", newToken)
				return next(w, r.WithContext(ctx), ps)
			} else {
				return apierrors.ErrUnauthorized
			}
		}

		return next(w, r.WithContext(ctx), ps)
	}
}

func (h *AuthServiceHandler) InternMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		claims := r.Context().Value("claims").(models.Claims)
		if claims.Role != "intern" {
			return apierrors.ErrForbidden
		}
		return next(w, r, ps)
	}
}

func (h *AuthServiceHandler) CompanyMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		claims := r.Context().Value("claims").(models.Claims)
		if claims.Role != "company" {
			return apierrors.ErrForbidden
		}
		return next(w, r, ps)
	}
}
