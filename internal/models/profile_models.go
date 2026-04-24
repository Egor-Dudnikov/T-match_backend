package models

import "time"

type Profile struct {
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	BirthDate  time.Time `json:"birth_date"`
	Location   string    `json:"location"`
	University string    `json:"university"`
	Degree     string    `json:"degree"`
	Bio        string    `json:"bio"`
	Experience string    `json:"experience"`
	Image      string    `json:"image"`
}
