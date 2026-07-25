// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) UpdateProfileHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.Profile](r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	err = h.service.UpdateStudentProfile(ctx, profile)
	return err
}

func (h *ServiceHandler) GetMyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.service.GetMyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.ProfileResponse](w, profile)
	return err
}

func (h *ServiceHandler) UpdateCompanyProfileHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.CompanyProfile](r)
	if err != nil {
		return err
	}
	err = h.service.UpdateCompanyProfile(r.Context(), profile)
	return err
}

func (h *ServiceHandler) GetMyCompanyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.service.GetMyCompanyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.CompanyProfileResponse](w, profile)
	return err
}

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

func (h *ServiceHandler) SetMyAvatarHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}

	file, info, err := r.FormFile("avatar")
	if err != nil {
		return apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	defer file.Close()
	url, err := h.service.SetMyAvatar(ctx, info, file, claims)
	if err != nil {
		return err
	}
	err = encodeJSON[string](w, url)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h ServiceHandler) GetAllSkills(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skills, err := h.service.GetAllSkills(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Skill](w, skills)
	return err
}

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

func (h ServiceHandler) DeleteInternSkillsHandler(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.service.DeleteInternSkills(r.Context(), skillIDs.ID)
	return err
}

func (h ServiceHandler) GetMyResponsesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	responses, err := h.service.GetMyResponses(ctx)
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Response](w, responses)
	return err
}

func (h ServiceHandler) SearchCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	query := r.URL.Query().Get("query")
	location := r.URL.Query().Get("location")
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchCompany{}

	filters.Query = h.parseAndSetString(query)
	filters.Location = h.parseAndSetString(location)

	filters.Offset = h.parseAndSetInt(offset)
	filters.Limit = h.parseAndSetInt(limit)

	res, err := h.service.SearchCompany(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.CompanyProfile](w, res)
	return err
}

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
