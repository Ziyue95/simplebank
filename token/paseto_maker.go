package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// PasetoMaker: struct of Paseto maker, which implements the token.Maker interface(./token/maker.go)
// Use Paseto version 2 and symmetric key algorithm to sign the tokens
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte // array of byte
}

// Define CreateToken & VerifyToken method for *PasetoMaker to implement Maker interface
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// create a new token payload:
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// generate encrypted token using paseto.Encrypt()
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	// decrypt token using maker.symmetricKey and stored into paylaod
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// check if payload is valid
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	// Paseto version 2 uses Chacha20 Poly1305 algorithm to encrypt the payload
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}
