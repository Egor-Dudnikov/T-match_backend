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

func (h *ServiceHandler) NewIntershipHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims := ctx.Value("claims").(models.Claims)
	internship, err := decodeJSON[models.Internship](r)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	internship.CompanyId = claims.UserID
	internshipID, err := h.authService.NewInternship(ctx, internship, claims.UserID)
	if err != nil {
		return err
	}
	encodeJSON[int](w, internshipID)
	return nil
}

func (h *ServiceHandler) GetInternshipByIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	internshipResp, err := h.authService.GetInternshipById(ctx, id)
	if err != nil {
		return err
	}
	err = encodeJSON[models.InternshipResponse](w, internshipResp)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONEncodeFailed, err)
	}
	return nil
}

func (h *ServiceHandler) UpdateInternshipHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}
	internship, err := decodeJSON[models.InternshipUpdate](r)
	internship.Id = id
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	err = h.authService.UpdateInternship(ctx, internship)
	return err

}

func getIdURL(ps httprouter.Params) (int, error) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}
	return id, nil
}

func (h *ServiceHandler) ArchivedInternshipHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}
	err = h.authService.ArchivedInternship(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) AddInternshipSkillsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	err = h.authService.AddInternshipSkills(ctx, skillIDs.Id, id)
	return err
}

func (h *ServiceHandler) DeleteInternshipSkillsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	err = h.authService.DeleteInternshipSkills(ctx, id, skillIDs.Id)
	return err
}

func (h *ServiceHandler) RespondInternshipHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	err = h.authService.RespondInternship(ctx, id)
	return err
}

func (h *ServiceHandler) GetInternshipResponses(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	responses, err := h.authService.GetInternshipResponses(ctx, id)
	if err != nil {
		return err
	}
	encodeJSON[[]models.Response](w, responses)
	return nil
}

func (h *ServiceHandler) SetResponseStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()

	id, err := getIdURL(ps)
	if err != nil {
		return err
	}

	resp, err := decodeJSON[models.ResponseRequest](r)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	err = h.authService.SetResponseStatus(ctx, id, resp.Status)
	return err
}

func (h *ServiceHandler) SearchInternshipHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	query := r.URL.Query().Get("query")
	location := r.URL.Query().Get("location")
	salaryMax := r.URL.Query().Get("salary_max")
	salaryMin := r.URL.Query().Get("salary_min")
	durationMin := r.URL.Query().Get("duration_min")
	durationMax := r.URL.Query().Get("duration_max")
	skills := r.URL.Query()["skills"]
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	filters := models.SearchInternship{}

	if len(query) != 0 {
		filters.Query = &query
	}

	if len(location) != 0 {
		filters.Location = &location
	}

	if len(salaryMax) != 0 {
		if val, err := strconv.Atoi(salaryMax); err == nil {
			filters.SalaryMax = &val
		}
	}

	if len(salaryMin) != 0 {
		if val, err := strconv.Atoi(salaryMin); err == nil {
			filters.SalaryMin = &val
		}
	}

	if len(durationMin) != 0 {
		if val, err := strconv.Atoi(durationMin); err == nil {
			filters.DurationMin = &val
		}
	}

	if len(durationMax) != 0 {
		if val, err := strconv.Atoi(durationMax); err == nil {
			filters.DurationMax = &val
		}
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

	if len(sort) != 0 {
		filters.Sort = &sort
	}

	if len(order) != 0 {
		if val, err := strconv.Atoi(order); err == nil {
			filters.Order = &val
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

	res, err := h.authService.SearchInternship(r.Context(), filters)
	if err != nil {
		return err
	}

	encodeJSON[[]models.Internship](w, res)
	return nil
}
