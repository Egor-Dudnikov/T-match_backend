// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import "time"

// Profile is the intern profile data.
type Profile struct {
	ID         *int       `json:"id,omitempty" validate:"omitempty"`
	UserID     *int       `json:"user_id,omitempty" validate:"omitempty"`
	FirstName  *string    `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string    `json:"last_name,omitempty" validate:"omitempty,max=100"`
	BirthDate  *time.Time `json:"birth_date,omitempty" validate:"omitempty"`
	CityID     *int       `json:"city_id,omitempty" validate:"omitempty,min=1"`
	University *string    `json:"university,omitempty" validate:"omitempty,max=200"`
	Degree     *string    `json:"degree,omitempty" validate:"omitempty,max=100"`
	Bio        *string    `json:"bio,omitempty" validate:"omitempty,max=2000"`
	Experience *string    `json:"experience,omitempty" validate:"omitempty,max=5000"`
	Image      *string    `json:"image,omitempty" validate:"omitempty"`
}

// CompanyProfile is the company profile data.
type CompanyProfile struct {
	ID           *int    `json:"id,omitempty" validate:"omitempty"`
	UserID       *int    `json:"user_id,omitempty" validate:"omitempty"`
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

// ProfileResponse is the intern profile returned by the API.
type ProfileResponse struct {
	Email   string  `json:"email,omitempty"`
	Profile Profile `json:"profile"`
	Skills  []Skill `json:"skills"`
}

// CompanyProfileResponse is the company profile returned by the API.
type CompanyProfileResponse struct {
	Email   string         `json:"email,omitempty"`
	Profile CompanyProfile `json:"profile"`
}

// Skill is a skill attached to a user or internship.
type Skill struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SkillID is the payload for adding or removing skills by id.
type SkillID struct {
	ID []int `json:"skill_id"`
}

// SearchCompany are the company search filters.
type SearchCompany struct {
	Query  *string
	CityID *int
	Limit  *int
	Offset *int
}

// SearchIntern are the student search filters.
type SearchIntern struct {
	Query      *string
	University *string
	Skills     *[]int
	Limit      *int
	Offset     *int
}

// ShortProfile is a trimmed profile used in search results.
type ShortProfile struct {
	ID         *int    `json:"id,omitempty" validate:"omitempty"`
	UserID     *int    `json:"user_id,omitempty" validate:"omitempty"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	CityID     *int    `json:"city_id,omitempty" validate:"omitempty,min=1"`
	University *string `json:"university,omitempty" validate:"omitempty,max=200"`
	Degree     *string `json:"degree,omitempty" validate:"omitempty,max=100"`
	Image      *string `json:"image,omitempty" validate:"omitempty"`
}

// City is a city with its geographic region.
type City struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

// Notification is a WebSocket notification sent to a user.
type Notification struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Type      string      `json:"type"`
	IsRead    bool        `json:"is_read"`
	CreatedAt time.Time   `json:"created_at"`
	Data      interface{} `json:"data,omitempty"`
}

// InvateData is the notification payload for an internship invite.
type InvateData struct {
	ID             int     `json:"id"`
	NotificationID int     `json:"notification_id"`
	InternshipID   int     `json:"internship_id"`
	CompanyID      int     `json:"company_id"`
	Message        *string `json:"message,omitempty"`
}

// ChangeStatusData is the notification payload for a response status change.
type ChangeStatusData struct {
	ID             int    `json:"id"`
	NotificationID int    `json:"notification_id"`
	InternshipID   int    `json:"internship_id"`
	CompanyID      int    `json:"company_id"`
	NewStatus      string `json:"new_status"`
}

// NewApplicationData is the notification payload for a new internship application.
type NewApplicationData struct {
	ID             int `json:"id"`
	NotificationID int `json:"notification_id"`
	InternshipID   int `json:"internship_id"`
	InternID       int `json:"intern_id"`
	ResponseID     int `json:"response_id"`
}
