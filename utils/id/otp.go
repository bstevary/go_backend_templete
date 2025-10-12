package id

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// secureRandomInt generates a cryptographically secure random integer in the range [0, max).
func secureRandomInt(max int64) (int32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return int32(n.Int64()), nil
}

func GenerateOTP() (string, error) {
	otp, err := secureRandomInt(1_000_000)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", otp), nil
}
