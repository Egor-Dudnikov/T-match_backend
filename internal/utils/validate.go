// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package utils

import (
	"regexp"

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
