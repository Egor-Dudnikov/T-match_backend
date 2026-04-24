package handlers

import (
	"T-match_backend/internal/models"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *AuthServiceHandler) UbdateProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.Profile](r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	err = h.authService.NewProfile(ctx, profile)
	if err != nil {
		return err
	}
	return nil
}
