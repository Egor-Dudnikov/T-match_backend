package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id           int
	Email        string
	Role         string
	PasswordHash string
}

type UserAuth struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
}

type UserVerify struct {
	Email        string
	PasswordHash string
	DeviceID     string
	Code         string
	Attempts     int
	CodeAmount   int
}

type CompanyAuth struct {
	Inn      string `json:"inn" validate:"required,min=10,max=12,numeric"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72,strong_password"`
	DeviceID string `json:"device_id" validate:"required,min=5,max=100"`
}

type CompanyVerify struct {
	CompanyData  CompanyData
	Email        string
	PasswordHash string
	DeviceID     string
	Code         string
	Attempts     int
	CodeAmount   int
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
