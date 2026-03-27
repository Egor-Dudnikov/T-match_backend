package utils

import (
	"crypto/rand"
	"math/big"
)

func NewCode() (string, error) {
	code := make([]byte, 6)
	max := big.NewInt(10)

	for i := range 6 {
		digit, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = byte('0' + digit.Int64())
	}
	return string(code), nil
}
