package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"db.sqlc.dev/app/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware: high-order authentication function, returns the authentication middleware function(gin.HandlerFunc)
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	// define the authentication middleware function
	return func(ctx *gin.Context) {
		// extract the authorization header from the request
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		// if authorizationHeader is empty
		if len(authorizationHeader) == 0 {
			err := errors.New("Authorization header is not provided")
			// abort the request by calling ctx.AbortWithStatusJSON() and send JSON response to client
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// authorizationHeader should have format starting with Bearer prefix + space + access token
		// Bearer prefix: let the server know the type of authorization
		// string.Fields(): split the authorization header by space
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// check authorization type
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// store the payload in the context
		ctx.Set(authorizationPayloadKey, payload)
		// forward the request to the next handler
		ctx.Next()
	}
}
