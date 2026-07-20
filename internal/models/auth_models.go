// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           int
	Email        string
	Role         string
	PasswordHash string
}

type InternAuth struct {
	Email     string    `json:"email" validate:"required,email,max=255"`
	Password  string    `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID  string    `json:"device_id" validate:"required,min=5,max=100"`
	BirthDate time.Time `json:"birth_date"`
}

type CompanyAuth struct {
	Inn      string `json:"inn" validate:"required,min=10,max=12,numeric"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
}

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

type InternVerify struct {
	Email        string
	PasswordHash string
	DeviceID     string
	BirthDate    time.Time
}

type CompanyVerify struct {
	CompanyData  CompanyData
	Email        string
	PasswordHash string
	DeviceID     string
}

type VerifyRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

type Claims struct {
	UserID   int
	DeviceID string
	Email    string
	Role     string
	jwt.RegisteredClaims
}

type UserInfo struct {
	UserID   int
	DeviceID string
	Email    string
	Role     string
}

type LoginUser struct {
	Email        string
	DeviceID     string
	PasswordHash string
}

type FogetPasswordRequest struct {
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Role     string `json:"role" validate:"required,valid_role"`
}

type ChangePasswordRequest struct {
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
}
