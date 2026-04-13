package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *AuthServiceHandler) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", ErrorMiddleware(
		app.CorsMiddleware(app.Index)))
	router.POST("/auth/students", ErrorMiddleware(
		app.CorsMiddleware(app.AuthStudentHandler)))
	router.POST("/auth/students/verify", ErrorMiddleware(
		app.CorsMiddleware(app.VerifyUserHandler)))
	router.POST("/auth/students/newverify", ErrorMiddleware(
		app.CorsMiddleware(app.NewVerifyCode)))
	router.POST("/auth/students/login", ErrorMiddleware(
		app.CorsMiddleware(app.LoginUserHandler)))

	router.OPTIONS("/auth/students", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/students/verify", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/students/newverify", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/students/login", ErrorMiddleware(app.CorsMiddleware(handleOptions)))

	return router
}

// заглушка для OPTIONS

func handleOptions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
