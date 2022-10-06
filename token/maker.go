package token

import (
	"time"
)

// Maker interface: manage the creation and verification of the tokens
type Maker interface {
	// CreateToken: returns a signed token string or an error
	CreateToken(username string, duration time.Duration) (string, error)
	// VerifyToken: checks if input token is valid and return the payload data stored inside the body of the token
	VerifyToken(token string) (*Payload, error)
}
