package unit_test

import (
	"testing"

	"phcmis/services/auth"
	"phcmis/test"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := test.RandomString(6)

	hashedPassword1, err := auth.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = auth.CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	wrongPassword := test.RandomString(6)
	err = auth.CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPassword2, err := auth.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
