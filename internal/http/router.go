package http

import (
	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *App) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", app.Index)
	router.POST("/auth/students", app.СreateUserHandler)
	return router
}
