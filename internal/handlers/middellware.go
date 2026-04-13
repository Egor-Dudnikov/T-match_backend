package handlers

import (
	"T-match_backend/internal/apierrors"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func ErrorMiddleware(next ErrorHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := next(w, r, ps); err != nil {
			status, message := apierrors.HTTPStatusMapping(err)
			http.Error(w, message, status)
			log.Println(err)
		}
	}
}

func (h *AuthServiceHandler) CorsMiddleware(next ErrorHandler) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(h.corsConfig.ControlAllowOrigin, ", "))
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.corsConfig.ControlAllowHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return next(w, r, ps)
	}
}
