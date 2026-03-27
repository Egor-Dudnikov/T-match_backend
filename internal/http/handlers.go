package http

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"time"

	"T-match_backend/internal/rw"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Db  *sql.DB
	Dbr *redis.Client
	Cfg rw.Config
}

func (app *App) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hi men"))
}

func (app *App) AuthStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userReg := rw.UserRegistration{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userReg)
	if err != nil {
		log.Println("JSON decoder error", err)
		return
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)

	sesionToken, err := rw.GeneratingJWT("0", userReg.Email, 3*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}

	code, err := NewCode()
	if err != nil {
		log.Println(err)
		return
	}

	err = rw.SendVerifyCode(userReg.Email, code, app.Cfg.VeryfyConfig)
	if err != nil {
		log.Println(err)
		return
	}

	user := rw.UserVerify{
		Email:        userReg.Email,
		PasswordHash: string(hashPassword),
		Code:         code,
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
	}

	if err := app.Dbr.Set(context.Background(), sesionToken, userJson, time.Minute*5).Err(); err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Token", sesionToken)
	log.Println("User registrathion")

}

func (app *App) VerifyStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sesionToken := r.Header.Get("Token")
	userVerify := rw.UserVerify{}
	res, err := app.Dbr.Get(context.Background(), sesionToken).Result()
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal([]byte(res), &userVerify)
	if err != nil {
		log.Println(err)
	}

	verifyRequest := rw.VerifyRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&verifyRequest)
	if err != nil {
		log.Println(err)
		return
	}

	if userVerify.Code != verifyRequest.Code {
		log.Println("Not right verify code")
		http.Error(w, "Not right verify code", http.StatusUnauthorized)
		return
	}

	user := rw.User{
		Email:        userVerify.Email,
		Role:         "intern",
		PasswordHash: userVerify.PasswordHash,
	}

	id, err := rw.QueeryNewUser(user, app.Db)
	user.Id = id
	if err != nil {
		log.Println("Error append database", err)
		return
	}

	accessToken, err := rw.GeneratingJWT(string(id), user.Email, time.Minute*7)
	refreshToken, err := rw.GeneratingJWT(string(id), user.Email, time.Hour*24*7)
	if err != nil {
		log.Println(err)
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
	if err := app.Dbr.Set(context.Background(), string(id), refreshToken, 7*24*time.Hour).Err(); err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Token", accessToken)
	log.Println("User registered successfully with id:", id)
}

func NewCode() (string, error) {
	code := make([]byte, 6)
	max := big.NewInt(10)

	for i := range 6 {
		digit, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = byte('0' + digit.Int64())
	}
	return string(code), nil
}
