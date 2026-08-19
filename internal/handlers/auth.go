// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

// ServiceHandler handles HTTP requests and delegates work to the underlying service.
type ServiceHandler struct {
	service      *service.Service
	corsConfig   *models.CORSConfig
	cookieSecure bool
}

// NewServiceHandler creates a new ServiceHandler with the given service, CORS config,
// and cookie security flag.
func NewServiceHandler(service *service.Service, cfg *models.CORSConfig, cookieSecure bool) *ServiceHandler {
	return &ServiceHandler{service: service, corsConfig: cfg, cookieSecure: cookieSecure}
}

// CheckHealth writes an "OK" response to indicate the service is healthy.
func (h *ServiceHandler) CheckHealth(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	_, err := w.Write([]byte("OK"))
	return err
}

// AuthStudentHandler registers a new intern and returns a verification session ID.
func (h *ServiceHandler) AuthStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userReg, err := decodeJSON[models.InternAuth](r)
	if err != nil {
		return err
	}

	sessionID, err := h.service.AuthIntern(ctx, userReg)
	if err != nil {
		return err
	}

	w.Header().Set("X-Verify-Session", sessionID)
	w.WriteHeader(http.StatusCreated)
	return nil
}

// VerifyUserHandler verifies an intern's session and returns access and refresh tokens.
func (h *ServiceHandler) VerifyUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("X-Verify-Session")
	if sessionID == "" {
		return apierrors.ErrBadRequest
	}

	verifyRequest, err := decodeJSON[models.VerifyRequest](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.VerifyUser(ctx, sessionID, verifyRequest, constants.Intern)
	if err != nil {
		return err
	}

	h.SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)

	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

// NewVerifyCode requests a new verification code for the given session.
func (h *ServiceHandler) NewVerifyCode(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("X-Verify-Session")
	err := h.service.NewCode(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// LoginUserHandler logs an intern in and returns access and refresh tokens.
func (h *ServiceHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.LoginUser](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.LoginUser(ctx, userLog, constants.Intern)
	if err != nil {
		return err
	}

	h.SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

// AuthCompanyHandler registers a new company and returns a verification session ID.
func (h *ServiceHandler) AuthCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userReg, err := decodeJSON[models.CompanyAuth](r)
	if err != nil {
		return err
	}

	sessionID, err := h.service.AuthCompany(ctx, userReg)
	if err != nil {
		return err
	}

	w.Header().Set("X-Verify-Session", sessionID)
	w.WriteHeader(http.StatusCreated)
	return nil
}

// VerifyCompanyHandler verifies a company's session and returns access and refresh tokens.
func (h *ServiceHandler) VerifyCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("X-Verify-Session")
	if sessionID == "" {
		return apierrors.ErrBadRequest
	}

	verifyRequest, err := decodeJSON[models.VerifyRequest](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.VerifyUser(ctx, sessionID, verifyRequest, constants.Company)
	if err != nil {
		return err
	}

	h.SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

// LoginCompanyHandler logs a company in and returns access and refresh tokens.
func (h *ServiceHandler) LoginCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.LoginUser](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.LoginUser(ctx, userLog, constants.Company)
	if err != nil {
		return err
	}

	h.SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

// LogoutHandler deletes the refresh token and clears the refresh cookie.
func (h *ServiceHandler) LogoutHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	err := h.service.DeleteRefreshToken(r.Context())
	if err != nil {
		return err
	}
	h.SetRefreshCookie(w, "", -1)
	return nil
}

// ForgotPasswordHandler starts a password reset flow and returns a verification session ID.
func (h *ServiceHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	req, err := decodeJSON[models.FogetPasswordRequest](r)
	if err != nil {
		return err
	}

	ctx := r.Context()
	sessionID, err := h.service.FogotPassword(ctx, req)
	if err != nil {
		return err
	}

	w.Header().Set("X-Verify-Session", sessionID)

	return nil
}

// VerifyForgotPasswordHandler verifies a password reset code and returns access and refresh tokens.
func (h *ServiceHandler) VerifyForgotPasswordHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("X-Verify-Session")
	if sessionID == "" {
		return apierrors.ErrBadRequest
	}

	verifyRequest, err := decodeJSON[models.VerifyRequest](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.VerifyFogottenUser(ctx, sessionID, verifyRequest)
	if err != nil {
		return err
	}

	h.SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

// ChangePasswordHandler changes the password for the authenticated user.
func (h *ServiceHandler) ChangePasswordHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	req, err := decodeJSON[models.ChangePasswordRequest](r)
	if err != nil {
		return err
	}

	ctx := r.Context()

	return h.service.ChangePassword(ctx, req)
}

func decodeJSON[T any](r *http.Request) (T, error) {
	var res T
	defer func() {
		if cerr := r.Body.Close(); cerr != nil {
			log.Printf("handlers: close request body: %v", cerr)
		}
	}()
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&res)
	if err != nil {
		return res, apierrors.Wrap(apierrors.ErrJSONDecodeFailed, err)
	}
	return res, nil
}

func encodeJSON[T any](w http.ResponseWriter, resp T) error {
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJSON)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrJSONEncodeFailed, err)
	}
	return nil
}

// SetRefreshCookie sets the refresh token cookie with the given value and max age.
func (h *ServiceHandler) SetRefreshCookie(w http.ResponseWriter, value string, maxAge int) {
	//nolint:gosec // Secure is driven by COOKIE_SECURE env: false for plain-HTTP dev, true in prod over HTTPS; the analyzer cannot see the runtime value.
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   maxAge,
	})
}
