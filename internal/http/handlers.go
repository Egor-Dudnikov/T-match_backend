package http

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"T-match_backend/internal/rw"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Db  *sql.DB
	Log *log.Logger
}

func (app *App) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hi men"))
}

func (app *App) AuthStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userReg := rw.UserRegistration{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userReg)
	if err != nil {
		app.Log.Println("JSON decoder error", err)
		return
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)

	user := rw.User{
		PasswordHash: string(hashPassword),
		Email:        userReg.Email,
		Role:         "intern",
	}

	id, err := rw.QueeryNewUser(user, app.Db)
	user.Id = id
	if err != nil {
		app.Log.Println("Error append database", err)
		return
	}

	accessToken, err := rw.GeneratingJWT(string(id), user.Email, time.Minute*7)
	refreshToken, err := rw.GeneratingJWT(string(id), user.Email, time.Hour*24*7)
	if err != nil {
		app.Log.Println(err)
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

	w.Header().Set("Token", accessToken)
	log.Println("User registered successfully with id:", id)
}

func (app *App) VerifyStudent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
