// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import "time"

type Profile struct {
	Id         *int       `json:"id,omitempty" validate:"omitempty"`
	FirstName  *string    `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string    `json:"last_name,omitempty" validate:"omitempty,max=100"`
	BirthDate  *time.Time `json:"birth_date,omitempty" validate:"omitempty"`
	Location   *string    `json:"location,omitempty" validate:"omitempty,max=200"`
	University *string    `json:"university,omitempty" validate:"omitempty,max=200"`
	Degree     *string    `json:"degree,omitempty" validate:"omitempty,max=100"`
	Bio        *string    `json:"bio,omitempty" validate:"omitempty,max=2000"`
	Experience *string    `json:"experience,omitempty" validate:"omitempty,max=5000"`
	Image      *string    `json:"image,omitempty" validate:"omitempty"`
}

type CompanyProfile struct {
	Id           *int    `json:"id,omitempty" validate:"omitempty"`
	CompanyName  *string `json:"company_name,omitempty" validate:"omitempty,max=100"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Website      *string `json:"website,omitempty" validate:"omitempty,max=200"`
	Inn          *string `json:"inn,omitempty"`
	Kpp          *string `json:"kpp,omitempty"`
	Ogrn         *string `json:"ogrn,omitempty"`
	LegalAddress *string `json:"legal_address,omitempty"`
	DirectorName *string `json:"director_name,omitempty"`
	Image        *string `json:"image,omitempty" validate:"omitempty"`
}

type ProfileResponse struct {
	Email   string  `json:"email,omitempty"`
	Profile Profile `json:"profile"`
	Skills  []Skill `json:"skills"`
}

type CompanyProfileResponse struct {
	Email   string         `json:"email,omitempty"`
	Profile CompanyProfile `json:"profile"`
}

type Skill struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type SkillID struct {
	Id []int `json:"skill_id"`
}

type SearchCompany struct {
	Query    *string
	Location *string
	Limit    *int
	Offset   *int
}

type SearchIntern struct {
	Query      *string
	University *string
	Skills     *[]int
	Limit      *int
	Offset     *int
}

type ShortProfile struct {
	Id         *int    `json:"id,omitempty" validate:"omitempty"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Location   *string `json:"location,omitempty" validate:"omitempty,max=200"`
	University *string `json:"university,omitempty" validate:"omitempty,max=200"`
	Degree     *string `json:"degree,omitempty" validate:"omitempty,max=100"`
	Image      *string `json:"image,omitempty" validate:"omitempty"`
}
