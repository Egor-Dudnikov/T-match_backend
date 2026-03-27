package handlers

import (
	"T-match_backend/internal/utils"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddelware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, claims, err := utils.DecodeJWT(authHeader)
		if err != nil {
			log.Println(err)
			http.Error(w, "User not autorization", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			refTocenCookie, err := r.Cookie("refresh_token")
			if err != nil {
				log.Println(err)
				http.Error(w, "User not autorization", http.StatusUnauthorized)
				return
			}
			var refTocen *jwt.Token
			refTocen, claims, err = utils.DecodeJWT(refTocenCookie.Value)
			if err != nil {
				log.Println(err)
				http.Error(w, "User not autorization", http.StatusUnauthorized)
				return
			}
			if !refTocen.Valid {
				log.Println("Unauthtorization")
				http.Error(w, "User not autorization", http.StatusUnauthorized)
				return
			} else {
				newAccessTocen, err := utils.GeneratingJWT(claims.UserID, claims.DeviceID, claims.Email, time.Hour*7*24)
				if err != nil {
					log.Println(err)
					http.Error(w, "User not autorization", http.StatusUnauthorized)
					return
				}
				w.Header().Set("New-Access-Token", newAccessTocen)
			}
		}
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
