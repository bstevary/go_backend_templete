package unit_test

import (
	"testing"
	"time"

	"phcmis/services/auth"
	"phcmis/test"

	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := auth.NewPasetoMaker(test.RandomString(32))
	require.NoError(t, err)

	username := test.RandomEmail()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, _, err := maker.CreateToken(username, duration, "")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.ValidateToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Email)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {

	maker, err := auth.NewPasetoMaker(test.RandomString(32))
	require.NoError(t, err)

	token, _, err := maker.CreateToken(test.RandomEmail(), -time.Minute, "")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.ValidateToken(token)
	require.Error(t, err)
	require.EqualError(t, err, auth.ErrExpiredToken.Error())
	require.Nil(t, payload)
}
