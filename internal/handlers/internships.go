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

func (h *ServiceHandler) NewInternshipHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims, ok := ctx.Value(constants.ClaimsKey).(models.Claims)
	if !ok {
		return apierrors.ErrInternalServer
	}
	internship, err := decodeJSON[models.Internship](r)
	if err != nil {
		return err
	}
	internship.CompanyID = claims.UserID
	internshipID, err := h.service.NewInternship(ctx, internship, claims.UserID)
	if err != nil {
		return err
	}
	err = encodeJSON[int](w, internshipID)
	w.WriteHeader(http.StatusCreated)
	return err
}

func (h *ServiceHandler) GetInternshipByIDHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	internshipResp, err := h.service.GetInternshipByID(ctx, id)
	if err != nil {
		return err
	}
	err = encodeJSON[models.InternshipResponse](w, internshipResp)
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) UpdateInternshipHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}
	internship, err := decodeJSON[models.InternshipUpdate](r)
	internship.ID = id
	if err != nil {
		return err
	}
	err = h.service.UpdateInternship(ctx, internship)
	return err

}

func getIDURL(ps httprouter.Params) (int, error) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, apierrors.Wrap(apierrors.ErrBadRequest, err)
	}
	return id, nil
}

func (h *ServiceHandler) ArchivedInternshipHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}
	err = h.service.ArchivedInternship(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) AddInternshipSkillsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.service.AddInternshipSkills(ctx, skillIDs.ID, id)
	w.WriteHeader(http.StatusCreated)
	return err
}

func (h *ServiceHandler) DeleteInternshipSkillsHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	skillIDs, err := decodeJSON[models.SkillID](r)
	if err != nil {
		return err
	}
	err = h.service.DeleteInternshipSkills(ctx, id, skillIDs.ID)
	return err
}

func (h *ServiceHandler) RespondInternshipHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	err = h.service.RespondInternship(ctx, id)
	return err
}

func (h *ServiceHandler) GetInternshipResponses(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	responses, err := h.service.GetInternshipResponses(ctx, id)
	if err != nil {
		return err
	}
	err = encodeJSON[[]models.Response](w, responses)
	return err
}

func (h *ServiceHandler) SetResponseStatus(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()

	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	resp, err := decodeJSON[models.ResponseRequest](r)
	if err != nil {
		return err
	}
	err = h.service.SetResponseStatus(ctx, id, resp.Status)
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

	res, err := h.service.SearchInternship(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.Internship](w, res)
	return err
}
