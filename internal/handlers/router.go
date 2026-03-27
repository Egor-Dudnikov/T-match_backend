package handlers

import (
	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *AuthServiceHandler) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/auth/students", app.AuthStudentHandler)
	router.POST("/auth/students/verify", app.VerifyStudentHandler)
	return router
}
