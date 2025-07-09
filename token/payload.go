package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ^ a custom error type
var ErrExpiredToken = errors.New("token has expired")

type Payload struct {
	Id        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {

	tokenId, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}
	// TODO : why is is referenced? -->
	//* Cheaper and faster when the struct grows beyond a couple of words.
	//*  exact object, not a copy.	Callers can change fields (e.g., set an expiry) and the change is visible to the callee and vice-versa.
	//*Lets you pass the payload to libraries that rely on those interfaces (e.g.,
	//*marsaller depnds on the &struct
	payload := &Payload{
		Id:        tokenId,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, err
}

func (payload *Payload) Valid() error {

	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
