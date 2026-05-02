package handlers

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *ServiceHandler) UbdateProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := decodeJSON[models.Profile](r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	err = h.authService.UpdateStudentProfile(ctx, profile)
	if err != nil {
		return err
	}
	return nil
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
	if err != nil {
		return err
	}
	return nil
}

func (h *ServiceHandler) GetMyCompanyProfileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	profile, err := h.authService.GetMyCompanyProfile(r.Context())
	if err != nil {
		return err
	}
	encodeJSON[models.CompanyProfileResponse](w, profile)
	return nil
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
