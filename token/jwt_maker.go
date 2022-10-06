package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTMaker: struct of JSON web token maker, which implements the token.Maker interface(./token/maker.go)
// Use symmetric key algorithm to sign the tokens
type JWTMaker struct {
	secretKey string
}

// Define CreateToken & VerifyToken method for *JWTMaker to implement Maker interface
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// create a new token payload:
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// create a new jwtToken by calling the jwt.NewWithClaims() function of the jwt-go package
	// jwt.SigningMethodHS256: the signing method(algorithm)
	// payload: the claims which implement jwt.Claims interface
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// generate a token string bycall jwtToken.SignedString()
	// pass in the secretKey after converting it to []byte
	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken: verify the tocken and return the payload data
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc: key function that receives the parsed but unverified token,
	// verify its header to make sure that the signing algorithm matches with what you use to sign the tokens
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// token.Method: get its signing algorithm(interface), and should convert it to a specific implementation(*jwt.SigningMethodHMAC)
		// *jwt.SigningMethodHMAC: struct containing HS256(currently used signing algorithm)
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		// the algorithm of the token doesn’t match with signing algorithm
		if !ok {
			return nil, ErrInvalidToken
		}
		// return the secret key if it matches
		return []byte(maker.secretKey), nil
	}

	// STEP 1.: parse the token by calling jwt.ParseWithClaims
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// possible cases: either the token is invalid or it is expired
		// convert returned error of ParseWithClaims() to jwt.ValidationError to figure out the real error type
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// STEP 2.: get payload data by converting jwtToken.Claims into a Payload object
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

// set the minimum size of secret key
const minSecretKeySize = 32

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}
