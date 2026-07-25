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

func (h *ServiceHandler) DeleteRespondInternshipHandler(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	id, err := getIDURL(ps)
	if err != nil {
		return err
	}

	err = h.service.DeleteRespondInternship(ctx, id)
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

	filters.Query = h.parseAndSetString(query)
	filters.Location = h.parseAndSetString(location)

	filters.SalaryMax = h.parseAndSetInt(salaryMax)
	filters.SalaryMin = h.parseAndSetInt(salaryMin)
	filters.DurationMin = h.parseAndSetInt(durationMin)
	filters.DurationMax = h.parseAndSetInt(durationMax)

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

	filters.Sort = h.parseAndSetString(sort)

	filters.Order = h.parseAndSetInt(order)
	filters.Offset = h.parseAndSetInt(offset)
	filters.Limit = h.parseAndSetInt(limit)

	res, err := h.service.SearchInternship(r.Context(), filters)
	if err != nil {
		return err
	}

	err = encodeJSON[[]models.Internship](w, res)
	return err
}

func (h *ServiceHandler) parseAndSetInt(value string) *int {
	if len(value) != 0 {
		if val, err := strconv.Atoi(value); err == nil {
			return &val
		}
	}
	return nil
}

func (h *ServiceHandler) parseAndSetString(val string) *string {
	if len(val) != 0 {
		return &val
	}
	return nil
}
