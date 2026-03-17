package rw

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GeneratingJWT(userID, email string, timeLife time.Duration) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	claims := Claims{
		userID,
		email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeLife)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "realword",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)

}

func DecodeJWT(tokenStr string) (*jwt.Token, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, keyfunc)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func keyfunc(token *jwt.Token) (interface{}, error) {
	return []byte(os.Getenv("JWT_SECRET")), nil
}
