package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"T-match_backend/internal/models"
	"T-match_backend/internal/repository"
	"T-match_backend/internal/service"
	"T-match_backend/internal/utils"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Db  *sql.DB
	Dbr *redis.Client
	Cfg models.Config
}

func (app *App) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hi men"))
}

func (app *App) AuthStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userReg := models.UserRegistration{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userReg)
	if err != nil {
		log.Println("JSON decoder error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	exist, err := repository.CheckUserEmail(userReg.Email, app.Db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exist {
		log.Println("User with this email exist")
		http.Error(w, "Conflict", http.StatusConflict)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sesionToken, err := utils.GeneratingJWT("0", userReg.DeviceID, userReg.Email, 3*time.Minute)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	code, err := utils.NewCode()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = service.SendVerifyCode(userReg.Email, code, app.Cfg.VeryfyConfig)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := models.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := app.Dbr.Set(context.Background(), sesionToken, userJson, time.Minute*5).Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Token", sesionToken)
	log.Println("User registrathion")

}

func (app *App) VerifyStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sesionToken := r.Header.Get("Token")
	if sesionToken == "" {
		log.Println("No token")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	userVerify := models.UserVerify{}
	res, err := app.Dbr.Get(context.Background(), sesionToken).Result()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	verifyRequest := models.VerifyRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&verifyRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if userVerify.Code != verifyRequest.Code {
		log.Println("Not right verify code")
		http.Error(w, "Not right verify code", http.StatusUnauthorized)
		return
	}

	app.Dbr.Del(context.Background(), sesionToken)

	_, claim, err := utils.DecodeJWT(sesionToken)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:        userVerify.Email,
		Role:         "intern",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := repository.QueeryNewUser(user, app.Db)
	user.Id = id
	if err != nil {
		log.Println("Error append database", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessToken, err := utils.GeneratingJWT(string(id), claim.DeviceID, user.Email, time.Minute*15)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GeneratingJWT(string(id), claim.DeviceID, user.Email, time.Hour*24*7)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})

	key := fmt.Sprintf("%d%s", id, claim.DeviceID)
	if err := app.Dbr.Set(context.Background(), key, refreshToken, 7*24*time.Hour).Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Token", accessToken)
	log.Println("User registered successfully with id:", id)
}
