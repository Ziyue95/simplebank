package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	accountFrom := createRandomAccount(t)
	accountTo := createRandomAccount(t)
	fmt.Println(">> before:", accountFrom.Balance, accountTo.Balance)

	n := 5
	amount := int64(10)

	// Channel is designed to connect concurrent Go routines,
	// and allow them to safely share data with each other without explicit locking.
	errs := make(chan error)
	results := make(chan TransferTxResult)

	// run n concurrent transfer transactions
	for i := 0; i < n; i++ {
		// txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			// pass the transaction name(txName) into the context
			//ctx := context.WithValue(context.Background(), txKey, txName)
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: accountFrom.ID,
				ToAccountID:   accountTo.ID,
				Amount:        amount,
			})

			// Inside the go routine, send err to the errs channel using this arrow operator <-
			// Then, we will check these errors and results from outside
			errs <- err
			results <- result
		}()
	}

	// Declare a new variable called existed of type map[int]bool
	// map[int]bool: map with int key and bool value;
	// make(map): build an empty map using make
	existed := make(map[int]bool)

	// check results
	for i := 0; i < n; i++ {
		// Receive the error from the errs channel
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// Check transfers
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, accountFrom.ID, transfer.FromAccountID)
		require.Equal(t, accountTo.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// Get transfers
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, accountFrom.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, accountTo.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts' balance
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, accountFrom.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, accountTo.ID, toAccount.ID)

		// check balances
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := accountFrom.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - accountTo.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balance
	updatedAccountFrom, err := store.GetAccount(context.Background(), accountFrom.ID)
	require.NoError(t, err)

	updatedAccountTo, err := store.GetAccount(context.Background(), accountTo.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccountFrom.Balance, updatedAccountTo.Balance)

	require.Equal(t, accountFrom.Balance-int64(n)*amount, updatedAccountFrom.Balance)
	require.Equal(t, accountTo.Balance+int64(n)*amount, updatedAccountTo.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	accountFrom := createRandomAccount(t)
	accountTo := createRandomAccount(t)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	// run n concurrent transfer transactions
	for i := 0; i < n; i++ {
		fromAccountID := accountFrom.ID
		toAccountID := accountTo.ID
		// switch the account ID when i%2==1
		if i%2 == 1 {
			fromAccountID = accountTo.ID
			toAccountID = accountFrom.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			// Inside the go routine, send err to the errs channel using this arrow operator <-
			// Then, we will check these errors and results from outside
			errs <- err
		}()
	}

	// check errs
	for i := 0; i < n; i++ {
		// Receive the error from the errs channel
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(context.Background(), accountFrom.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), accountTo.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, accountFrom.Balance, updatedAccount1.Balance)
	require.Equal(t, accountTo.Balance, updatedAccount2.Balance)
}
