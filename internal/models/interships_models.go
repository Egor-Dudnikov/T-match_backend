// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package models

import "time"

type Internship struct {
	ID            int       `json:"id"`
	CompanyID     int       `json:"company_id"`
	Title         string    `json:"title" validate:"required,max=200"`
	Description   string    `json:"description" validate:"required,max=5000"`
	Salary        int       `json:"salary" validate:"omitempty,min=0"`
	Location      string    `json:"location" validate:"required,max=200"`
	IsArchived    bool      `json:"is_archived"`
	DurationMonth int       `json:"duration_months" validate:"required,min=1"`
	CreatedAt     time.Time `json:"created_at"`
}

type InternshipUpdate struct {
	ID            int     `json:"id,omitempty"`
	Title         *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description   *string `json:"description,omitempty" validate:"omitempty,min=1,max=5000"`
	Salary        *int    `json:"salary,omitempty" validate:"omitempty,min=0"`
	Location      *string `json:"location,omitempty" validate:"omitempty,min=1,max=200"`
	DurationMonth *int    `json:"duration_months,omitempty" validate:"omitempty,min=1"`
}

type InternshipResponse struct {
	Internship Internship `json:"internship"`
	Skills     []Skill    `json:"skills"`
}

type Response struct {
	ID           int       `json:"id"`
	InternID     int       `json:"intern_id"`
	InternshipID int       `json:"internship_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type ResponseRequest struct {
	Status string `json:"status" validate:"required,valid_response_status"`
}

type SearchInternship struct {
	Query       *string
	Location    *string
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

type InvateIntern struct {
	UserID  int     `json:"user_id" validate:"required,numeric"`
	Message *string `json:"message" validate:"omitempty"`
}
