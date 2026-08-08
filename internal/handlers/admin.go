// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) LoginAdminHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.LoginUser](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.service.LoginUser(ctx, userLog, constants.Admin)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken, constants.MaxAgeRefreshToken)
	err = encodeJSON(w, map[string]string{"access_token": accessToken})
	return err
}

func (h *ServiceHandler) AdminGetStatsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	stats, err := h.service.GetAdminStats(r.Context())
	if err != nil {
		return err
	}

	err = encodeJSON[models.AdminStats](w, stats)
	return err
}

func (h *ServiceHandler) AdminBanUserHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	userID, err := getIDURL(ps)
	if err != nil {
		return err
	}

	req, err := decodeJSON[models.AdminBanRequest](r)
	if err != nil {
		return err
	}

	err = h.service.BanUser(r.Context(), userID, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *ServiceHandler) AdminUnbanUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	userID, err := getIDURL(ps)
	if err != nil {
		return err
	}

	err = h.service.UnbanUser(r.Context(), userID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *ServiceHandler) AdminDeleteInternshipHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	internshipID, err := getIDURL(ps)
	if err != nil {
		return err
	}

	err = h.service.AdminDeleteInternship(r.Context(), internshipID)
	if err != nil {
		return err
	}

	return nil
}
