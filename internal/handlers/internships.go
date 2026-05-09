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
	encodeJSON[[]models.Response](w, responses)
	return nil
}
