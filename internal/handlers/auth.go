// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"encoding/json"
	"net/http"

	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type ServiceHandler struct {
	service    *service.Service
	corsConfig *models.CORSConfig
}

func NewServiceHandler(service *service.Service, cfg *models.CORSConfig) *ServiceHandler {
	return &ServiceHandler{service: service, corsConfig: cfg}
}

func (h *ServiceHandler) CheckHealth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	w.Write([]byte("OK"))
	return nil
}

func (h *ServiceHandler) AuthStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userReg, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}

	sessionID, err := h.service.AuthUser(ctx, userReg)
	if err != nil {
		return err
	}

	w.Header().Set("X-Verify-Session", sessionID)
	w.WriteHeader(http.StatusCreated)
	return nil
}

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
	accessToken, refreshToken, err := h.service.VerifyUser(ctx, sessionID, verifyRequest)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)

	encodeJSON(w, map[string]string{"access_token": accessToken})
	return nil
}

func (h *ServiceHandler) NewVerifyCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("X-Verify-Session")
	err := h.service.NewCode(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.LoginUser(ctx, userLog)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	encodeJSON(w, map[string]string{"access_token": accessToken})
	return nil
}

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
	accessToken, refreshToken, err := h.service.VerifyCompany(ctx, sessionID, verifyRequest)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	encodeJSON(w, map[string]string{"access_token": accessToken})
	return nil
}

func (h *ServiceHandler) LoginCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.LoginCompany(ctx, userLog)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	encodeJSON(w, map[string]string{"access_token": accessToken})
	return nil
}

func (h *ServiceHandler) LogoutHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	h.service.DeleteRefreshToken(r.Context())
	SetRefreshCookie(w, "", -1)
	return nil
}

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

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	encodeJSON(w, map[string]string{"access_token": accessToken})
	return nil
}

func (h *ServiceHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	req, err := decodeJSON[models.ChangePasswordRequest](r)
	if err != nil {
		return err
	}

	ctx := r.Context()

	return h.service.ChangePassword(ctx, req)
}

func decodeJSON[T any](r *http.Request) (T, error) {
	var res T
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&res)
	if err != nil {
		return res, apierrors.Warp(apierrors.ErrJSONDecodeFailed, err)
	}
	return res, nil
}

func encodeJSON[T any](w http.ResponseWriter, resp T) error {
	respJson, err := json.Marshal(resp)
	if err != nil {
		return apierrors.Warp(apierrors.ErrJSONEncodeFailed, err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respJson)
	if err != nil {
		return apierrors.Warp(apierrors.ErrJSONEncodeFailed, err)
	}
	return nil
}

func SetRefreshCookie(w http.ResponseWriter, value string, maxAge int) {
	// при переходе на https заменить Secure на true
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   maxAge,
	})
}
