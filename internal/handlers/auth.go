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
}

func NewAuthServiceHandler(authService *service.AuthService) *AuthServiceHandler {
	return &AuthServiceHandler{authService: authService}
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hi men"))
}

func (h *AuthServiceHandler) AuthStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	userReg := models.UserAuth{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&userReg)

	if err != nil {
		return fmt.Errorf("%W: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	sessionID, err := h.authService.AuthUser(userReg)
	if err != nil {
		return err
	}

	w.Header().Set("Token", sessionID)
	log.Println("User registrathion successfully")
	return nil
}

func (h *AuthServiceHandler) VerifyStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	sessionID := r.Header.Get("Token")
	if sessionID == "" {
		return fmt.Errorf(apierrors.ErrBadRequest.Error())
	}

	verifyRequest := models.VerifyRequest{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&verifyRequest)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	accessToken, refreshToken, err := h.authService.VerifyStudent(sessionID, verifyRequest)
	if err != nil {
		return err
	}
	// при переходе на https заменить Secure на true
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Token", accessToken)
	log.Println("User verifycation successfully")
	return nil
}

func (h *AuthServiceHandler) NewVerifyCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	sessionID := r.Header.Get("Token")
	err := h.authService.NewCode(sessionID)
	if err != nil {
		return err
	}
	return nil
}

func (h *AuthServiceHandler) LoginStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	userLog := models.UserAuth{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&userLog)
	if err != nil {
		return fmt.Errorf("%w: %v", apierrors.ErrJSONDecodeFailed, err)
	}

	accessToken, refreshToken, err := h.authService.LoginStudent(userLog)
	if err != nil {
		return err
	}

	// при переходе на https заменить Secure на true
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Token", accessToken)
	log.Println("User verifycation successfully")
	return nil
}
