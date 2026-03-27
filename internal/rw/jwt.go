package rw

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GeneratingJWT(userID, deviceID, email string, timeLife time.Duration) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	claims := Claims{
		userID,
		deviceID,
		email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeLife)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "t-match_backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)

}

func DecodeJWT(tokenStr string) (*jwt.Token, Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, keyfunc)
	if err != nil {
		return nil, claims, err
	}
	return token, claims, nil
}

func keyfunc(token *jwt.Token) (interface{}, error) {
	return []byte(os.Getenv("JWT_SECRET")), nil
}
