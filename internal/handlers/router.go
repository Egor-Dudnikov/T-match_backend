package handlers

import (
	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *AuthServiceHandler) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/auth/students", ErrorMiddelware(app.AuthStudentHandler))
	router.POST("/auth/students/verify", ErrorMiddelware(app.VerifyStudentHandler))
	router.POST("/auth/students/newverify", ErrorMiddelware(app.NewVerifyCode))
	router.POST("/auth/students/login", ErrorMiddelware(app.LoginStudentHandler))
	return router
}
