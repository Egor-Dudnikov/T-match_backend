// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import "time"

// Internship is an internship listing published by a company.
type Internship struct {
	ID            int       `json:"id"`
	CompanyID     int       `json:"company_id"`
	Title         string    `json:"title" validate:"required,max=200"`
	Description   string    `json:"description" validate:"required,max=5000"`
	Salary        int       `json:"salary" validate:"omitempty,min=0"`
	CityID        int       `json:"city_id" validate:"required,min=1"`
	IsArchived    bool      `json:"is_archived"`
	DurationMonth int       `json:"duration_months" validate:"required,min=1"`
	CreatedAt     time.Time `json:"created_at"`
}

// InternshipUpdate is the payload for updating an internship.
type InternshipUpdate struct {
	ID            int     `json:"id,omitempty"`
	Title         *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description   *string `json:"description,omitempty" validate:"omitempty,min=1,max=5000"`
	Salary        *int    `json:"salary,omitempty" validate:"omitempty,min=0"`
	CityID        *int    `json:"city_id,omitempty" validate:"omitempty,min=1"`
	DurationMonth *int    `json:"duration_months,omitempty" validate:"omitempty,min=1"`
}

// InternshipResponse is the internship returned by the API along with its skills.
type InternshipResponse struct {
	Internship Internship `json:"internship"`
	Skills     []Skill    `json:"skills"`
}

// Response is a student's application to an internship.
type Response struct {
	ID           int       `json:"id"`
	InternID     int       `json:"intern_id"`
	InternshipID int       `json:"internship_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ResponseRequest is the payload for changing a response status.
type ResponseRequest struct {
	Status string `json:"status" validate:"required,valid_response_status"`
}

// SearchInternship are the internship search filters.
type SearchInternship struct {
	Query       *string
	CityID      *int
	SalaryMax   *int
	SalaryMin   *int
	DurationMin *int
	DurationMax *int
	Skills      *[]int
	Sort        *string
	Order       *int
	Offset      *int
	Limit       *int
}

// InvateIntern is the payload for inviting a student to an internship.
type InvateIntern struct {
	UserID  int     `json:"user_id" validate:"required,numeric"`
	Message *string `json:"message" validate:"omitempty"`
}
