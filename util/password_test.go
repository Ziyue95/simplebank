package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomString(6)

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	// generate wrongPassword and expect mismatch
	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashedPassword1)
	// use require.EqualError() to compare the output error. It must be equal to the bcrypt.ErrMismatchedHashAndPassword error
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	// test if the same password is hashed twice, 2 different hash values should be produced
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
