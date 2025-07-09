package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/itsadijmbt/simple_bank/db/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.Id)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

// ^Using maker.CreateToken() correctly signed JWT that’s just already expired
func TestExpiredJWTToken(t *testing.T) {

	maker, err := NewJWTMaker(util.RandomString(32))

	require.NoError(t, err)

	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	// NoError here cause we want no error while tocken creation
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	// we want a "Error"  not a "NoError"
	require.Error(t, err)
	// we want to match the error
	require.EqualError(t, err, ErrExpiredToken.Error())
	// we want our payload to be nil
	require.Nil(t, payload)
}

// ^ TestInvalidJWTTokenAlgNone ensures that tokens signed with "alg: none"
// ^ We want a forged token with no signature, so we bypass maker.CreateToken()
func TestInvalidJWTTokenAlgNone(t *testing.T) {
	// Step 1: Create a valid payload with a short expiry
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)

	// Step 2: Create a forged token using 'none' algorithm (no signature)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)

	// Step 3: Generate token string, explicitly allowing unsafe 'none' signature type
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Step 4: Create a JWT maker instance with a random secret key
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	// passing a forged token (with alg: none) to  real, secure JWT verifier (maker.VerifyToken).
	// That verifier will parse the token, inspect its alg header, and call  keyfunc.

	payload, err = maker.VerifyToken(token)

	require.Error(t, err)                               // an error must be returned
	require.EqualError(t, err, ErrInvalidToken.Error()) // specifically ErrInvalidToken
	require.Nil(t, payload)                             // payload must not be returned
}
