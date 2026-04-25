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
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthStudentHandler, 20, "/auth/students"))))

	router.POST("/auth/students/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyUserHandler, 60, "/auth/students/verify"))))

	router.POST("/auth/newverify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.NewVerifyCode, 7, "/auth/newverify"))))

	router.POST("/auth/students/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginUserHandler, 30, "/auth/students/login"))))

	router.POST("/auth/company", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthCompanyHandler, 20, "/auth/company"))))

	router.POST("/auth/company/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyCompanyHandler, 60, "/auth/company/verify"))))

	router.POST("/auth/company/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginCompanyHandler, 30, "/auth/company/login"))))

	router.PUT("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.RateLimitMiddleware(app.UbdateProfile, 300, "/my/profile")))))

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
