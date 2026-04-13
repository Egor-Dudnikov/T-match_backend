package handlers

import (
	"T-match_backend/internal/apierrors"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func ErrorMiddelware(next ErrorHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := next(w, r, ps); err != nil {
			status, message := apierrors.HTTPStatusMapping(err)
			http.Error(w, message, status)
			log.Println(err)
		}
	}
}

func CorsMiddelware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return next(w, r, ps)
	}
}
