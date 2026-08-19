// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package utils

import (
	"T-match_backend/internal/constants"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ValidPassword reports whether the field value is a valid password.
func ValidPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var hasDigit bool
	var hasUpperLetter bool
	var hasLowerLetter bool

	for _, r := range password {
		if unicode.Is(unicode.Latin, r) && unicode.IsUpper(r) {
			hasUpperLetter = true
		} else if unicode.Is(unicode.Latin, r) && unicode.IsLower(r) {
			hasLowerLetter = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		} else if !isAllowSpecial(r) {
			return false
		}
	}

	return hasDigit && hasLowerLetter && hasUpperLetter
}

func isAllowSpecial(c rune) bool {
	for _, r := range constants.AllowSpecialPassword {
		if c == r {
			return true
		}
	}
	return false
}

// ValidRole reports whether the field value is a valid role.
func ValidRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	if role != constants.Intern && role != constants.Company {
		return false
	}
	return true
}

// ValidAge reports whether the person born on the given date is old enough.
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

// ValidResponseStatus reports whether the field value is a valid response status.
func ValidResponseStatus(fl validator.FieldLevel) bool {
	statuses := [4]string{constants.Pending, constants.Reviewing, constants.Accepted, constants.Rejected}
	status := fl.Field().String()
	ok := false
	for _, st := range statuses {
		if status == st {
			ok = true
			break
		}
	}
	return ok
}

// NewValidator creates a validator with the registered validation rules.
func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	err := validate.RegisterValidation("strong_password", ValidPassword)
	if err != nil {
		return validate, err
	}

	err = validate.RegisterValidation("valid_role", ValidRole)
	if err != nil {
		return validate, err
	}

	err = validate.RegisterValidation("valid_response_status", ValidResponseStatus)
	if err != nil {
		return validate, err
	}

	return validate, nil
}
