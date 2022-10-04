package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store interface should have all functions of the Queries struct,
// and one more function to execute the transfer money transaction
type Store interface {
	// include all query methods in Querier interface(./db/sqlc/querier.go)
	Querier
	// add func TransferTx to enable money transfer between accounts
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore is a concrete type that have methods required by Store interface
// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries         // SQLStore extent the Queries object to support db transactions
	db       *sql.DB // required to create a new db transaction
}

// NewStore creates a new Store interface
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction, unexported function -- cannot called by external packages
// input: context & a callback function of created Queries object -> returns error
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {

	// start a new transaction with provided level of isolation(nil for default(read-committed))
	// BeginTx will return tx -- a transaction object(sql.tx) if error is nil
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// pass tx to New() function to create the queries that runs within transaction
	q := New(tx)
	// check the error message return by callback function fn after the queries
	err = fn(q)
	// rollback the transaction if err is not nil: call tx.Rollback()
	if err != nil {
		// call tx.Rollback() and store rollback err into rbErr
		// if If the rollback error is also not nil, then we have to report 2 errors
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		// only return err if rollback is successful
		return err
	}

	// commit the transaction if err is nil: call tx.Commit()
	return tx.Commit()
}

// The TransferTxParams struct contains all necessary input parameters to transfer money between 2 accounts
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// The TransferTxResult struct contains the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// declare variable txKey of type empty(struct{})
// second {} means creating a new empty object of that type(struct{})
// var txKey = struct{}{}

// TransferTx performs the money transfer transaction
// It creates a transfer record, add account entries, and update accounts' balance within a single db transaction
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// implement the callback function: use queries object q to call individual CRUD function
		var err error

		// get transaction name from context
		// txName := ctx.Value(txKey)

		// fmt.Println(txName, "create transfer")
		// step 1. create transfer and return err if err != nil
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// step 2. create the FromEntry and ToEntry and return err if err != nil
		// fmt.Println(txName, "create fromEntry")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "create toEntry")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// step 3. update the FromAccount and ToAccount and return err if err != nil
		// It involves locking and preventing potential deadlock
		// Avoid deadlock by making sure the account with smaller ID is updated first
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return err
	})

	// return the result and the error of the execTx() call
	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
