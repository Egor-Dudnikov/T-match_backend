package handlers

import (
	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *AuthServiceHandler) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/auth/students", ErrorMidelware(app.AuthStudentHandler))
	router.POST("/auth/students/verify", ErrorMidelware(app.VerifyStudentHandler))
	return router
}
