package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/models"
	"T-match_backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type AuthServiceHandler struct {
	authService *service.AuthService
	corsConfig  *models.CORSConfig
}

func NewAuthServiceHandler(authService *service.AuthService) *AuthServiceHandler {
	return &AuthServiceHandler{authService: authService}
}

func (h *AuthServiceHandler) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	w.Write([]byte("Hi men"))
	return nil
}

func (h *AuthServiceHandler) AuthStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userReg, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}

	sessionID, err := h.authService.AuthUser(ctx, userReg)
	if err != nil {
		return err
	}

	w.Header().Set("Token", sessionID)
	log.Println(sessionID)
	return nil
}

func (h *AuthServiceHandler) VerifyUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("Token")
	if sessionID == "" {
		return fmt.Errorf("%w", apierrors.ErrBadRequest)
	}

	verifyRequest, err := decodeJSON[models.VerifyRequest](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.authService.VerifyUser(ctx, sessionID, verifyRequest)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken)

	w.Header().Set("Token", accessToken)
	log.Println(sessionID)
	return nil
}

func (h *AuthServiceHandler) NewVerifyCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("Token")
	err := h.authService.NewCode(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}

func (h *AuthServiceHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.authService.LoginUser(ctx, userLog)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken)
	w.Header().Set("Token", accessToken)
	return nil
}

func (h *AuthServiceHandler) AuthCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userReg, err := decodeJSON[models.CompanyAuth](r)
	if err != nil {
		return err
	}

	sessionID, err := h.authService.AuthCompany(ctx, userReg)
	if err != nil {
		return err
	}

	w.Header().Set("Token", sessionID)
	log.Println(sessionID)
	return nil
}

func (h *AuthServiceHandler) VerifyCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	sessionID := r.Header.Get("Token")
	if sessionID == "" {
		return fmt.Errorf("%w", apierrors.ErrBadRequest)
	}

	verifyRequest, err := decodeJSON[models.VerifyRequest](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.authService.VerifyCompany(ctx, sessionID, verifyRequest)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken)

	w.Header().Set("Token", accessToken)
	log.Println(sessionID)
	return nil
}

func (h *AuthServiceHandler) LoginCompanyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	ctx := r.Context()
	userLog, err := decodeJSON[models.UserAuth](r)
	if err != nil {
		return err
	}
	accessToken, refreshToken, err := h.authService.LoginCompany(ctx, userLog)
	if err != nil {
		return err
	}

	SetRefreshCookie(w, refreshToken)
	w.Header().Set("Token", accessToken)
	return nil
}

func decodeJSON[T any](r *http.Request) (T, error) {
	var res T
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&res)
	if err != nil {
		return res, fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}
	return res, nil
}

func SetRefreshCookie(w http.ResponseWriter, value string) {
	// при переходе на https заменить Secure на true
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   604800,
	})
}
