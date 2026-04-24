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
	router.POST("/auth/newverify", ErrorMiddleware(
		app.CorsMiddleware(app.NewVerifyCode)))
	router.POST("/auth/students/login", ErrorMiddleware(
		app.CorsMiddleware(app.LoginUserHandler)))
	router.POST("/auth/company", ErrorMiddleware(
		app.CorsMiddleware(app.AuthCompanyHandler)))
	router.POST("/auth/company/verify", ErrorMiddleware(
		app.CorsMiddleware(app.VerifyCompanyHandler)))
	router.POST("/auth/company/login", ErrorMiddleware(
		app.CorsMiddleware(app.LoginCompanyHandler)))
	router.PUT("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(app.UbdateProfile)))))

	router.OPTIONS("/auth/students", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/students/verify", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/students/login", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/company", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/company/verify", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/newverify", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/auth/company/login", ErrorMiddleware(app.CorsMiddleware(handleOptions)))
	router.OPTIONS("/my/profile", ErrorMiddleware(app.CorsMiddleware(handleOptions)))

	return router
}

// заглушка для OPTIONS

func handleOptions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
