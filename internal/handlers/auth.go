package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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

func (h *AuthServiceHandler) AuthStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userReg := models.UserRegistration{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userReg)
	if err != nil {
		log.Println("JSON decoder error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	sesionToken, err := h.authService.AuthUser(userReg)

	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Token", sesionToken)
	log.Println("User registrathion")
}

func (h *AuthServiceHandler) VerifyStudentHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sesionToken := r.Header.Get("Token")
	if sesionToken == "" {
		log.Println("No token")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	verifyRequest := models.VerifyRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&verifyRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := h.authService.VerifyStudent(sesionToken, verifyRequest)

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Token", accessToken)
	log.Println("User verifycation successfully")
}
