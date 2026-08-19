// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

// Package utils provides shared helpers for JWT, validation and code generation.
package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

// NewCode generates a random six-digit verification code.
func NewCode() (string, error) {
	code := make([]byte, 0, 6)
	maxInt := big.NewInt(10)

	for range 6 {
		digit, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			return "", err
		}
		code = strconv.AppendInt(code, digit.Int64(), 10)
	}
	return string(code), nil
}
