package handlers

import (
	"T-match_backend/internal/models"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) NewIntershipHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	claims := ctx.Value("claims").(models.Claims)
	internship, err := decodeJSON[models.Internship](r)
	if err != nil {
		return err
	}
	internship.CompanyId = claims.UserID
	err = h.authService.NewInternship(ctx, internship, claims.UserID)
	if err != nil {
		return err
	}
	return nil
}
