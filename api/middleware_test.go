package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"db.sqlc.dev/app/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// addAuthorization: create a new access token, and add it to the authorization header of the request
// given: authorizationType, username, duration
func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	// create token
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	// create authorization header value with string format(authorizationType token)
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	// add header to request
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	// table-driven test strategy
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			// happy case:
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
		{
			// client does not provide any kind of authorization header
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No addAuthorization
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// change authorizationType to an unsupported one
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			// client does not provide authorization type prefix
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// change authorizationType to an unsupported one
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// change authorizationType to an unsupported one
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recoder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// don't need to access the Store -> set store argument as nil
			server := newTestServer(t, nil)
			// add a simple API route and handler to test the middleware
			authPath := "/auth"
			// declare the route with server.router.GET
			server.router.GET(
				authPath,
				// create authMiddleware with server.tokenMaker and add it to the route
				authMiddleware(server.tokenMaker),
				// add the handler function: simple send a status 200 OK with empty body to the client
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			// send request to this API
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			// call tc.setupAuth to add authorization header to the request
			tc.setupAuth(t, request, server.tokenMaker)
			// serve the request and record in recorder
			server.router.ServeHTTP(recorder, request)
			// verify the response of request using tc.checkResponse
			tc.checkResponse(t, recorder)
		})
	}
}
