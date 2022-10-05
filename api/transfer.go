package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "db.sqlc.dev/app/db/sqlc"
	"github.com/gin-gonic/gin"
)

// similar struct as createAccountRequest in ./api/account.go
type transferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	// gt=0: require Amount to be greater than 0
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,currency"`
}

// validAccount checks if an account with a specific ID really exists, and its currency matches the input currency
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	// STEP 1.: check if input data is in valid format;
	// If the error is not nil, then it means that the client has provided invalid data
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// http.StatusBadRequest is the 404 HTTP status code
		// errorResponse(err) is the JSON object(error info) that we want to send to the client
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // call ctx.JSON() function to send a JSON response
		return
	}

	// STEP 2.: check if accounts exist and match the currency
	if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	// STEP 3.: insert the new transfer into the database;
	// TransferTxParams struct defined in ./db/store.go
	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// Function TransferTx defined in ./db/store.go
	result, err := server.store.TransferTx(ctx, arg)
	// send JSON response with 500 Internal Server Error status code to client if err is not nil
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// send a 200 OK status code to client if no error;
	ctx.JSON(http.StatusOK, result)

}
