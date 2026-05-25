// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) UbdateProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.Profile](r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	err = h.authService.UpdateStudentProfile(ctx, profile)
	return err
}

func (h *ServiceHandler) GetMyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.authService.GetMyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.ProfileResponse](w, profile)
	return err
}

func (h *ServiceHandler) UpdateCompanyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.CompanyProfile](r)
	if err != nil {
		return err
	}
	err = h.authService.UpdateCompanyProfile(r.Context(), profile)
	return err
}

func (h *ServiceHandler) GetMyCompanyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.authService.GetMyCompanyProfile(r.Context())
	if err != nil {
		return err
	}
	err = encodeJSON[models.CompanyProfileResponse](w, profile)
	return err
}

func (h *ServiceHandler) SetMyAvatarHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims := ctx.Value("claims").(models.Claims)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}

	file, info, err := r.FormFile("avatar")
	defer file.Close()
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrBadRequest, err)
	}
	url, err := h.authService.SetMyAvatar(ctx, info, file, claims)
	if err != nil {
		return err
	}
	err = encodeJSON[string](w, url)
	return err
}

func (h ServiceHandler) GetAllSkills(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skills, err := h.authService.GetAllSkills(r.Context())
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
	err = h.authService.AddInternSkills(r.Context(), skillIDs.Id)
	return err
}

func (h ServiceHandler) DeleteInternSkillsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.authService.DeleteInternSkills(r.Context(), skillIDs.Id)
	return err
}

func (h ServiceHandler) GetMyResponsesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	responses, err := h.authService.GetMyResponses(ctx)
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Response](w, responses)
	return err
}

func (h ServiceHandler) SerchCompanyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	query := r.URL.Query().Get("query")
	location := r.URL.Query().Get("location")
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchCompany{}

	if len(query) != 0 {
		filters.Query = &query
	}

	if len(location) != 0 {
		filters.Location = &location
	}

	if len(offset) != 0 {
		if val, err := strconv.Atoi(offset); err == nil {
			filters.Offset = &val
		}
	}

	if len(limit) != 0 {
		if val, err := strconv.Atoi(limit); err == nil {
			filters.Limit = &val
		}
	}
	res, err := h.authService.SearchCompany(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.Company](w, res)
	return err
}

func (h ServiceHandler) SerchInternHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	query := r.URL.Query().Get("query")
	university := r.URL.Query().Get("university")
	skills := r.URL.Query()["skills"]
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchIntern{}

	if len(query) != 0 {
		filters.Query = &query
	}

	if len(university) != 0 {
		filters.University = &university
	}

	if len(skills) != 0 {
		var skillIds []int
		for _, s := range skills {
			if id, err := strconv.Atoi(s); err == nil {
				skillIds = append(skillIds, id)
			}
		}
		if len(skillIds) > 0 {
			filters.Skills = &skillIds
		}
	}

	if len(offset) != 0 {
		if val, err := strconv.Atoi(offset); err == nil {
			filters.Offset = &val
		}
	}

	if len(limit) != 0 {
		if val, err := strconv.Atoi(limit); err == nil {
			filters.Limit = &val
		}
	}
	res, err := h.authService.SearchIntern(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.Intern](w, res)
	return err
}
