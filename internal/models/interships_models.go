package models

import "time"

type Internship struct {
	Id            int       `json:"id"`
	CompanyId     int       `json:"company_id"`
	Title         string    `json:"title" validate:"required,max=200"`
	Description   string    `json:"description" validate:"required,max=5000"`
	Salary        int       `json:"salary" validate:"omitempty,min=0"`
	Location      string    `json:"location" validate:"required,max=200"`
	IsArchived    bool      `json:"is_archived"`
	DurationMonth int       `json:"duration_month" validate:"required,min=1"`
	CreatedAt     time.Time `json:"created_at"`
}

type InternshipUpdate struct {
	Id            int    `json:"id" validate:"required"`
	Title         string `json:"title,omitempty" validate:"omitempty,max=200"`
	Description   string `json:"description,omitempty" validate:"omitempty,max=5000"`
	Salary        *int   `json:"salary,omitempty" validate:"omitempty,min=0"`
	Location      string `json:"location,omitempty" validate:"omitempty,max=200"`
	DurationMonth int    `json:"duration_month,omitempty" validate:"omitempty,min=1"`
}
