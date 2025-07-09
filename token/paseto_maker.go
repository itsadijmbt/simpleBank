package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto *paseto.V2
	//^ local system
	symmetricKey []byte
}

// new maker creates a new PasetoMaker
// !Initialize PasetoMaker ONCE per application
func NewPasetoMaker(symmetricKey string) (Maker, error) {

	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("ivalid key size , keysize should be %d", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// passing in the username and how long the token should be valid.
	payload, err := NewPayload(username, duration)

	if err != nil {
		return "", err
	}
	//* payload is serialized (converted to JSON) and encrypted with the secret key.
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {

	// ^ Always create a new empty struct to hold the decrypted payload.
	// ^ This makes the code safe, idempotent, and reusable across tokens.
	// ^ A memory address (pointer) to fill with data.
	// ^ A known struct type so it can map JSON fiel
	payload := &Payload{}

	//* paseto internally uses a dst is a pointer to struct, where the decrypted JSON will be unmarshaled
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)

	if err != nil {
		return nil, ErrInvalidToken
	}
	// validate payload fields (e.g., check token hasn't expired)
	err = payload.Valid()
	if err != nil {
		return nil, err
	}
	return payload, err

}

/*
*Function	  Role

?PasetoMaker	    Reusable engine to encrypt/decrypt using the key
?CreateToken()	    Makes new payload, encrypts it
?VerifyToken()  	Takes token, uses empty payload to decode
?Payload      		Struct that holds data inside the token (username, times, ID)

*Lifecycle

  At application startup:
 Call NewPasetoMaker() once to create and store the token engine with secret key.

  When user logs in:
 Call CreateToken(username, duration)
 → A new Payload is created
 → Encrypted using Paseto V2 and symmetric key
 → Returns a secure token string

  When user makes a request with token:
 Call VerifyToken(token)
 → Create an empty Payload struct (&Payload{})
 → Decrypt token using symmetric key
 → Paseto fills the empty payload with data from token
 → Validates timestamps (issue time, expiry time)
 → Returns the payload for further use


*/
