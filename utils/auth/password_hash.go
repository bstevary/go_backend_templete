package auth

//  this file is basic bcrypt implementation kept for reference but
// actual implementation is in pswrd_hash_argon2id.go using argon2id
// use this for backward compatibility if needed

// import (
// 	"fmt"

// 	"golang.org/x/crypto/bcrypt"
// )

// func HashPassword(password string) (string, error) {
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to hash password %w", err)
// 	}
// 	return string(hashedPassword), nil
// }

// func CheckPassword(password, hashedPassword string) error {
// 	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
// }
