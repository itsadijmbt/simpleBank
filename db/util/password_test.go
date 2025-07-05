package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {

	password := RandomString(10)

	hashedPass, err := HashedPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, hashedPass)

	err = CheckPassword(password, hashedPass)

	require.NoError(t, err)

	wrongPass := RandomString(10)
	//we creaye a wrong password and check if it's hash is also eq
	err = CheckPassword(wrongPass, hashedPass)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	//we wnat to esnure that 2hashed password are diff
	//i.e for a same password different hashes should be there
	// while in the above we check if 2 diff passes have same hashes

	hashedPass2, err := HashedPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, hashedPass2)

	require.NotEqual(t, hashedPass, hashedPass2)

}
