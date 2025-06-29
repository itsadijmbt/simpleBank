package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions.
type SQLStore struct {
	*Queries
	db *sql.DB
}

// ! interface for Gomocks
// ! the store interface has all the fucntions of the *quries struct + funciton to transfer money
type Store interface {
	//* now adding all fucnion of queries struct is difficukt so sqlc has emit_interface
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// NewStore creates a new Store.
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// txKeyType is a private type to avoid context-key collisions.
// type txKeyType struct{}

// var txKey = txKeyType{}

// execTx executes the given function within a database transaction.
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction.
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction.
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another.
// It creates a transfer record, two ledger entries, and updates both accounts’ balances.
// We also log the per-transaction name from the context for debugging.
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		// txName := ctx.Value(txKey)

		// 1) create transfer record

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		// 2) create debit entry

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// 3) create credit entry

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// 4) update balances (with SELECT ... FOR UPDATE inside GetAccountForUpdate)

		//^ IN OUR CASE TO HAVE A CONSISTENT DB AND DEADLOCK AVOIDANCE WE USE QUERYSEQUENCING := USE A ORDER OF TRANSC HERE WE USE SMALLER ID FIRST

		if arg.FromAccountID < arg.ToAccountID {

			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)

		} else {
			//^ to account should be updated!!

			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, +arg.Amount, arg.FromAccountID, -arg.Amount)

		}

		return nil
	})

	return result, err
}

//* why the traditional locking failed

//^ 1-> You never lock the rows you read, so between your SELECT and your UPDATE, another transfer can slip in and stomp on your balance calculation
//^ 2-> you must explicitly acquire locks when you need serialized access to specific rows.

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
		//^ here since we are using named return params we can just return empty!!
		// return account1, account2, err
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})

	if err != nil {
		return
		//^ here since we are using named return params we can just return empty!!
		// return account1, account2, err
	}

	return

}
