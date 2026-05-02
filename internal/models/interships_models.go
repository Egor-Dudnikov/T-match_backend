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
	DurationMonth int       `json:"duration_month" validate:"required,min=1,max=12"` // исправлено DurationMonth
	CreatedAt     time.Time `json:"created_at"`
}
