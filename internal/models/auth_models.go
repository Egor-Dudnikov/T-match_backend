// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User is a registered account in the system.
type User struct {
	ID           int
	Email        string
	Role         string
	PasswordHash string
}

// InternAuth is the registration payload for interns.
type InternAuth struct {
	Email     string    `json:"email" validate:"required,email,max=255"`
	Password  string    `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID  string    `json:"device_id" validate:"required,min=5,max=100"`
	BirthDate time.Time `json:"birth_date"`
}

// CompanyAuth is the registration payload for companies.
type CompanyAuth struct {
	Inn      string `json:"inn" validate:"required,min=10,max=12,numeric"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
}

// CompanyData holds the company details returned by the dadata service.
type CompanyData struct {
	Inn          string
	Kpp          string
	Ogrn         string
	Okved        string
	BranchType   string
	ShortName    string
	Status       string
	Director     string
	DirectorPost string
	Address      string
}

// InternVerify is the verification payload for interns.
type InternVerify struct {
	Email        string
	PasswordHash string
	DeviceID     string
	BirthDate    time.Time
}

// CompanyVerify is the verification payload for companies.
type CompanyVerify struct {
	CompanyData  CompanyData
	Email        string
	PasswordHash string
	DeviceID     string
}

// VerifyRequest is the email verification code payload.
type VerifyRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

// Claims is the JWT claims structure embedded into tokens.
type Claims struct {
	UserID   int
	DeviceID string
	Email    string
	Role     string
	jwt.RegisteredClaims
}

// UserInfo contains the identity data extracted from the access token.
type UserInfo struct {
	UserID   int
	DeviceID string
	Email    string
	Role     string
}

// LoginUser is the login payload.
type LoginUser struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password"`
	DeviceID     string `json:"device_id"`
}

// FogetPasswordRequest is the forgot-password request payload.
type FogetPasswordRequest struct {
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Role     string `json:"role" validate:"required,valid_role"`
}

// ChangePasswordRequest is the password change payload.
type ChangePasswordRequest struct {
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
}
