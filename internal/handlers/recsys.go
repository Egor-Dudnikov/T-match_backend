// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/models"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) GetRecommendationsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	recommendations, err := h.service.GetRecommendations(r.Context())
	if err != nil {
		return err
	}
	return encodeJSON[[]models.Recommendation](w, recommendations)
}

func (h *ServiceHandler) TrackInternshipViewHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	internshipID, err := getIDURL(ps)
	if err != nil {
		return err
	}
	h.service.TrackInternshipView(r.Context(), internshipID)
	return nil
}
