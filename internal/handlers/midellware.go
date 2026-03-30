package handlers

import (
	"T-match_backend/internal/apierrors"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func ErrorMidelware(next ErrorHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := next(w, r, ps); err != nil {
			status, message := apierrors.HTTPStatusMapping(err)
			http.Error(w, message, status)
			log.Println(err)
		}
	}
}
