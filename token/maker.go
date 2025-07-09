package token

import "time"

//* maker is an interface for token management
type Maker interface {

	//*token creation for a specific user and time
	CreateToken(username string, duration time.Duration) (string, error)

	//* returns the payload inside the token!
	VerifyToken(token string) (*Payload, error)
}
