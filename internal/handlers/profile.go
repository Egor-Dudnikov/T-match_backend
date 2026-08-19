// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"T-match_backend/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// UpdateProfileHandler updates the authenticated intern's profile.
func (h *ServiceHandler) UpdateProfileHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.Profile](r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	err = h.service.UpdateStudentProfile(ctx, profile)
	return err
}

// GetMyProfileHandler returns the authenticated intern's profile.
func (h *ServiceHandler) GetMyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.service.GetMyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.ProfileResponse](w, profile)
	return err
}

// UpdateCompanyProfileHandler updates the authenticated company's profile.
func (h *ServiceHandler) UpdateCompanyProfileHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.CompanyProfile](r)
	if err != nil {
		return err
	}
	err = h.service.UpdateCompanyProfile(r.Context(), profile)
	return err
}

// GetMyCompanyProfileHandler returns the authenticated company's profile.
func (h *ServiceHandler) GetMyCompanyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.service.GetMyCompanyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.CompanyProfileResponse](w, profile)
	return err
}

// GetCompanyProfileHandler returns a company profile by ID.
func (h *ServiceHandler) GetCompanyProfileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}
	profile, err := h.service.GetCompanyProfile(r.Context(), id)
	if err != nil {
		return err
	}
	err = encodeJSON[models.CompanyProfileResponse](w, profile)
	return err
}

// SetMyAvatarHandler uploads an avatar image for the authenticated user.
func (h *ServiceHandler) SetMyAvatarHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxSizeImage)
	err := r.ParseMultipartForm(constants.MaxSizeImage) //nolint:gosec // body is already capped by MaxBytesReader above; gosec cannot detect it
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	file, info, err := r.FormFile("avatar")
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("handlers: close avatar file: %v", cerr)
		}
	}()
	url, err := h.service.SetMyAvatar(ctx, info, file, claims)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	return encodeJSON[string](w, url)
}

// GetAllSkills returns all available skills.
func (h ServiceHandler) GetAllSkills(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skills, err := h.service.GetAllSkills(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Skill](w, skills)
	return err
}

// GetAllCities returns all available cities.
func (h ServiceHandler) GetAllCities(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	cities, err := h.service.GetAllCities(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.City](w, cities)
	return err
}

// AddInternSkillsHandler adds skills to the authenticated intern's profile.
func (h ServiceHandler) AddInternSkillsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.service.AddInternSkills(r.Context(), skillIDs.ID)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	return nil
}

// DeleteInternSkillsHandler removes skills from the authenticated intern's profile.
func (h ServiceHandler) DeleteInternSkillsHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.service.DeleteInternSkills(r.Context(), skillIDs.ID)
	return err
}

// GetMyResponsesHandler returns the authenticated intern's responses to internships.
func (h ServiceHandler) GetMyResponsesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	responses, err := h.service.GetMyResponses(ctx)
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Response](w, responses)
	return err
}

// SearchCompanyHandler searches companies by the provided filters.
func (h ServiceHandler) SearchCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	query := r.URL.Query().Get("query")
	cityID := r.URL.Query().Get("city_id")
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchCompany{}

	filters.Query = h.parseAndSetString(query)
	filters.CityID = h.parseAndSetInt(cityID)

	filters.Offset = h.parseAndSetInt(offset)
	filters.Limit = h.parseAndSetInt(limit)

	res, err := h.service.SearchCompany(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.CompanyProfile](w, res)
	return err
}

// SearchInternHandler searches interns by the provided filters.
func (h ServiceHandler) SearchInternHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	query := r.URL.Query().Get("query")
	university := r.URL.Query().Get("university")
	skills := r.URL.Query()["skills"]
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchIntern{}

	filters.Query = h.parseAndSetString(query)
	filters.University = h.parseAndSetString(university)

	if len(skills) != 0 {
		var skillIDs []int
		for _, s := range skills {
			if id, err := strconv.Atoi(s); err == nil {
				skillIDs = append(skillIDs, id)
			}
		}
		if len(skillIDs) > 0 {
			filters.Skills = &skillIDs
		}
	}

	filters.Offset = h.parseAndSetInt(offset)
	filters.Limit = h.parseAndSetInt(limit)

	res, err := h.service.SearchIntern(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.ShortProfile](w, res)
	return err
}

// GetProfileHandler returns a profile by ID.
func (h *ServiceHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}
	resp, err := h.service.GetProfile(ctx, id)
	if err != nil {
		return err
	}
	err = encodeJSON[models.ProfileResponse](w, resp)
	return err
}

// MyNotificationsHandler returns the authenticated user's notifications.
func (h *ServiceHandler) MyNotificationsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	notifications, err := h.service.GetMyNotifications(ctx)
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Notification](w, notifications)
	return err
}

// SetReadStatusOfNotificationHandler marks the authenticated user's notifications as read.
func (h *ServiceHandler) SetReadStatusOfNotificationHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	err := h.service.SetReadStatusOfNotification(ctx)
	return err
}

// WSNotificationHandler upgrades the connection to a WebSocket for real-time notifications.
func (h *ServiceHandler) WSNotificationHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  constants.WSReadBufferSize,
		WriteBufferSize: constants.WSWriteBufferSize,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	claims, ok := r.Context().Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}

	cli := service.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan string, constants.MaxBufferNotificationWS),
		Hub:    h.service.Hub,
	}

	h.service.Hub.Register(claims.UserID, &cli)

	go cli.WritePump()
	go cli.ReadPump()

	return nil
}
