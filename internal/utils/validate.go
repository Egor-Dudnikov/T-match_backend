// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package utils

import (
	"T-match_backend/internal/constants"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	hasUpper = regexp.MustCompile(`[A-Z]`)
	hasDigit = regexp.MustCompile(`[0-9]`)
	hasLower = regexp.MustCompile(`[a-z]`)
)

func ValidPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	has1 := hasUpper.MatchString(password)
	has2 := hasDigit.MatchString(password)
	has3 := hasLower.MatchString(password)

	return has1 && has2 && has3
}

func ValidRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	if role != constants.Intern && role != constants.Company {
		return false
	}
	return true
}

func ValidAge(birthDate time.Time) bool {
	today := time.Now()
	age := today.Year() - birthDate.Year()
	if today.YearDay() < birthDate.YearDay() {
		age--
	}
	if age < constants.MinUserAge {
		return false
	}
	return true
}
