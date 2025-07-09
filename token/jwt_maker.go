package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const minSecretKeySize = 32

// ^ JWT MAKER
type JWTMaker struct {
	secretKey string
}

// ^ NEWJWTMaker creates a new JWTMaker
var ErrInvalidToken = errors.New("token is Invalid")

func NewJWTMaker(secretKey string) (Maker, error) {

	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key szie and must be %d size", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {

	payload, err := NewPayload(username, duration)

	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))

}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {

	// Prevents algorithm substitution attacks.
	// Ensures only expected algorithm (HS256) is used.
	// Library passes token header here so you can use `kid` (key ID) if needed.
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is HMAC as it as type asserion on interface
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			// If not, reject this token as invalid (possibly tampered)
			return nil, ErrInvalidToken
		}

		// Return the expected secret key to verify the signature
		return []byte(maker.secretKey), nil
	}

	// Base64-decodes header and payload
	// Safer and more flexible than Parse() (which uses MapClaims)
	// Uses `keyfunc` to get the key and verify the signature
	// Populates the claims into the provided struct (&Payload{})
	// Passing &Payload{} allows the library to fill the struct with decoded claim values.
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyfunc)

	// We want to detect whether the error is specifically due to expiry (graceful fail),
	// or a more serious issue like signature mismatch or tampering.
	if err != nil {
		// Try to unwrap the validation error
		//extracting from the err interrface
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {

			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	//Type assertion claims to your custom *Payload type
	// The .Claims field is of type jwt.Claims (interface{})
	// So we must convert it back to our concrete type.
	// Defensive check to ensure our expected struct is used.
	// Protects against runtime panics in future access.
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {

		return nil, ErrInvalidToken
	}

	return payload, nil
}

//level diagram for refernce
/*
ParseWithClaims(tokenString, &Payload{}, keyfunc)
      │
      └───▶ parses token → builds jwt.Token object
                   │
                   └───▶ calls keyfunc(jwtToken)
                             │
                             └───▶ your func inspects token.Method and returns secret key
*/
