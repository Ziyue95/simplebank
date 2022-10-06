package token

import (
	"time"

	"github.com/google/uuid"
)

// Payload struct: contain the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// Maker interface: manage the creation and verification of the tokens
type Maker interface {
	// CreateToken: returns a signed token string or an error
	CreateToken(username string, duration time.Duration) (string, error)
	// VerifyToken: checks if input token is valid and return the payload data stored inside the body of the token
	VerifyToken(token string) (*Payload, error)
}
