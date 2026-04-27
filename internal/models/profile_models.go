package models

import "time"

type Profile struct {
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
	CompanyName *string `json:"company_name"`
	Description *string `json:"descrption"`
	Website     *string `json:"website"`
}
