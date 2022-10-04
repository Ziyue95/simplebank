package api

import (
	"database/sql"
	"net/http"

	db "db.sqlc.dev/app/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Balance of newly created account is 0
// Owner and currency info is obtained from the body of HTTP request(JSON object)
type createAccountRequest struct {
	// Use a binding tag to tell Gin that the field is required,
	// Later call ShouldBindJSON function to parse the input data from HTTP request body,
	// and Gin will validate the output object to make sure it satisfy the conditions we specified in the binding tag.
	Owner string `json:"owner" binding:"required"`
	// use the oneof condition to declare bank only supports 2 types of currency for now: USD and EUR
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
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

	// STEP 2.: insert the new account into the database;
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	// send JSON response with 500 Internal Server Error status code to client if err is not nil
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// send a 200 OK status code to client if no error;
	ctx.JSON(http.StatusOK, account)

}

type getAccountRequest struct {
	// use the uri tag to tell Gin the name of the URI paramete
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

	ctx.JSON(http.StatusOK, account)
}
