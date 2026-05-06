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
	err = h.authService.NewInternship(ctx, internship, claims.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) GetInternshipByIdHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	ctx := r.Context()
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrInternalServer, err)
	}
	internship, err := h.authService.GetInternshipById(ctx, id)
	if err != nil {
		return err
	}
	err = encodeJSON[models.Internship](w, internship)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONEncodeFailed, err)
	}
	return nil
}
