package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *AuthServiceHandler) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", ErrorMiddelware(
		CorsMiddelware(app.Index)))
	router.POST("/auth/students", ErrorMiddelware(
		CorsMiddelware(app.AuthStudentHandler)))
	router.POST("/auth/students/verify", ErrorMiddelware(
		CorsMiddelware(app.VerifyUserHandler)))
	router.POST("/auth/students/newverify", ErrorMiddelware(
		CorsMiddelware(app.NewVerifyCode)))
	router.POST("/auth/students/login", ErrorMiddelware(
		CorsMiddelware(app.LoginUserHandler)))

	router.OPTIONS("/auth/students", ErrorMiddelware(CorsMiddelware(handleOptions)))
	router.OPTIONS("/auth/students/verify", ErrorMiddelware(CorsMiddelware(handleOptions)))
	router.OPTIONS("/auth/students/newverify", ErrorMiddelware(CorsMiddelware(handleOptions)))
	router.OPTIONS("/auth/students/login", ErrorMiddelware(CorsMiddelware(handleOptions)))

	return router
}

// заглушка для OPTIONS

func handleOptions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
