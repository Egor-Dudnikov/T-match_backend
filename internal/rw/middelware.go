package rw

import (
	"net/http"
)

func AuthMiddelware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
