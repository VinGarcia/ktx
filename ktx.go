package ktx

import (
	"context"
	"database/sql"
	"fmt"
)

// DBRunner represents the minimal interface needed to execute database operations.
// It is compatible with database/sql standard library interfaces.
type DBRunner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// TxBeginner represents a database connection that can begin transactions.
type TxBeginner interface {
	DBRunner
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Tx represents a database transaction.
type Tx interface {
	DBRunner
	Rollback() error
	Commit() error
}

// Transaction encapsulates several database operations into a single transaction.
// All database operations should be performed inside the input callback `fn`
// using the provided DBRunner.
//
// If the callback returns any errors, the transaction will be rolled back,
// otherwise the transaction will be committed.
//
// If a panic occurs during the callback execution, the transaction will be
// rolled back and the panic will be re-raised.
//
// If the provided db is already a transaction (sql.Tx), it will be reused
// without starting a new transaction.
func Transaction(ctx context.Context, db DBRunner, fn func(DBRunner) error) error {
	// Check if db is already a transaction
	if tx, ok := db.(*sql.Tx); ok {
		return fn(tx)
	}

	// Check if db can begin transactions
	txBeginner, ok := db.(TxBeginner)
	if !ok {
		return fmt.Errorf("provided db does not implement TxBeginner interface")
	}

	// Start a new transaction
	tx, err := txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Handle panics by rolling back the transaction
	defer func() {
		if r := recover(); r != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				r = fmt.Errorf(
					"unable to rollback after panic with value: %v, rollback error: %w",
					r, rollbackErr,
				)
			}
			panic(r)
		}
	}()

	// Execute the callback with the transaction
	err = fn(tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			err = fmt.Errorf(
				"unable to rollback after error: %s, rollback error: %w",
				err, rollbackErr,
			)
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
