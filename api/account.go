package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "db.sqlc.dev/app/db/sqlc"
	"db.sqlc.dev/app/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Balance of newly created account is 0
// Owner and currency info is obtained from the body of HTTP request(JSON object)
type createAccountRequest struct {
	// Use a binding tag to tell Gin that the field is required,
	// Later call ShouldBindJSON function to parse the input data from HTTP request body,
	// and Gin will validate the output object to make sure it satisfy the conditions we specified in the binding tag.
	// Owner string `json:"owner" binding:"required"`

	// use the oneof condition to declare bank only supports 2 types of currency for now: USD and EUR
	// substitue oneof condition by custom currency validator
	Currency string `json:"currency" binding:"required,currency"`
}

// declare a function createAccount with a server pointer receiver
// createAccount requires a *gin.Context object as input to be consistent with handler function required by POST method
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// STEP 1.: check if input data is in valid format;
	// If the error is not nil, then it means that the client has provided invalid data
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// http.StatusBadRequest is the 404 HTTP status code
		// errorResponse(err) is the JSON object(error info) that we want to send to the client
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // call ctx.JSON() function to send a JSON response
		return
	}

	// Owner should be the username of the logged in user stored in the authorization payload:
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// STEP 2.: insert the new account into the database;
	arg := db.CreateAccountParams{
		// API RULE: A logged-in user can only create an accounr for him/herself
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	// send JSON response with 500 Internal Server Error status code to client if err is not nil
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// send a 200 OK status code to client if no error;
	ctx.JSON(http.StatusOK, account)

}

type getAccountRequest struct {
	// use the uri tag to tell Gin the name of the URI parameter
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// API RULE: A logged-in user can only get accounts that he/she owns
	if account.Owner != authPayload.Username {
		err := errors.New("account does not belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	// to get parameters from query string, use form tag
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.ListAccountsParams{
		// API RULE: A logged-in user can only list accounts that he/she owns
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
