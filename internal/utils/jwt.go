// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package utils

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GeneratingJWT creates a signed JWT with the given claims and lifetime.
func GeneratingJWT(userID int, deviceID, email, role string, timeLife time.Duration) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	claims := models.Claims{
		UserID:   userID,
		DeviceID: deviceID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeLife)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "t-match_backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		return tokenStr, apierrors.Wrap(apierrors.ErrJWTGenerationFailed, err)
	}
	return tokenStr, nil

}

// GeneratingTokenPair creates an access and refresh token pair for the given user.
func GeneratingTokenPair(userID int, deviceID, email, role string) (string, string, error) {
	accessToken, err := GeneratingJWT(userID, deviceID, email, role, constants.AccessTokenTimeLife)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := GeneratingJWT(userID, deviceID, email, role, constants.RefreshTokenTimeLife)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// DecodeJWT parses and verifies the given JWT, returning its token and claims.
func DecodeJWT(tokenStr string) (*jwt.Token, models.Claims, error) {
	claims := models.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, keyfunc)
	if err != nil {
		return token, claims, apierrors.Wrap(apierrors.ErrJWTDecodingFailed, err)
	}
	return token, claims, nil
}

func keyfunc(_ *jwt.Token) (interface{}, error) {
	return []byte(os.Getenv("JWT_SECRET")), nil
}
