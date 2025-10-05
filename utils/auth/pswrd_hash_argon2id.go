package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2ID is a password hashing algorithm.
type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
	SaltLength  uint32
}

var params = &Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 4,
	KeyLength:   32,
	SaltLength:  16,
}

func HashPassword(password string) (string, error) {

	// Generate a cryptographically secure random salt.
	salt := make([]byte, params.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("error generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Encode the salt and hashed password as base64.
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// Concatenate the salt and hashed password.
	encodedPassword := fmt.Sprintf("%s.%s", encodedSalt, encodedHash)

	return encodedPassword, nil
}

// Argon2ID is a password hashing algorithm.
func CheckPassword(hashedPassword, password string) error {
	// Split the encoded password into the salt and hashed password.
	parts := strings.Split(hashedPassword, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid hashed password")
	}

	// Decode the salt and hashed password.
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("error decoding salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("error decoding hash: %w", err)
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Compare the hashed password with the comparison hash.
	if !bytes.Equal(hash, comparisonHash) {
		return fmt.Errorf("password does not match")
	}

	return nil
}
