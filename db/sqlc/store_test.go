package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n, amount := 10, int64(10)
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		//& capture a fresh copy each iteration
		//& each goroutine now passes a fresh copy and does not use the last loop ID's
		fromID, toID := account1.ID, account2.ID
		if i%2 == 1 {
			fromID, toID = account2.ID, account1.ID
		}

		go func(f, t int64) {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: f,
				ToAccountID:   t,
				Amount:        amount,
			})
			errs <- err
		}(fromID, toID)
	}

	// collect errors
	for i := 0; i < n; i++ {
		require.NoError(t, <-errs)
	}

	// verify balances untouched
	updated1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updated2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updated1.Balance, updated2.Balance)
	require.Equal(t, account1.Balance, updated1.Balance)
	require.Equal(t, account2.Balance, updated2.Balance)
}

func TestTransferTx(t *testing.T) {
	// assume testStore, testQueries and createRandomAccount(t) are already
	// set up in your TestMain (in another _test.go)
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	n := 10
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// 1) fire off n concurrent transfers
	for i := 0; i < n; i++ {
		// txName := fmt.Sprintf("tx %d", i+1)
		//! Closure capture: we pass txName into the goroutine so each worker logs its own name.
		go func() {
			ctx := context.Background()
			res, err := testStore.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- res
		}()
	}

	// 2) collect & assert
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		require.NoError(t, <-errs)

		result := <-results
		require.NotEmpty(t, result)

		// transfer record
		tr := result.Transfer
		require.Equal(t, account1.ID, tr.FromAccountID)
		require.Equal(t, account2.ID, tr.ToAccountID)
		require.Equal(t, amount, tr.Amount)
		require.NotZero(t, tr.ID)

		// entries
		fromE := result.FromEntry
		require.Equal(t, account1.ID, fromE.AccountID)
		require.Equal(t, -amount, fromE.Amount)

		toE := result.ToEntry
		require.Equal(t, account2.ID, toE.AccountID)
		require.Equal(t, amount, toE.Amount)

		// **updated** accounts
		fromA := result.FromAccount
		require.Equal(t, account1.ID, fromA.ID)
		toA := result.ToAccount
		require.Equal(t, account2.ID, toA.ID)

		// both balances shift by the same delta
		diff1 := account1.Balance - fromA.Balance
		diff2 := toA.Balance - account2.Balance
		require.Equal(t, diff1, diff2)

		// diff1 should be a multiple of amount, between 1*amount and n*amount
		require.True(t, diff1 > 0)
		require.Zero(t, diff1%amount)
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		existed[k] = true
	}

	// ensure we saw both orderings (k == 1 and k == 2)
	require.Len(t, existed, n)

	// 3) final DB check with the same global testQueries
	updated1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updated2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updated1.Balance, updated2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updated1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updated2.Balance)
}
