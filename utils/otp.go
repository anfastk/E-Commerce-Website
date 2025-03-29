package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP(length int) string {
	if length <= 0 {
		return ""
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil) 
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%0*d", length, n)
} 