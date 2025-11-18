package ktx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestTransaction_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	err := Transaction(ctx, db, func(tx DBRunner) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Jane", "jane@example.com")
		return err
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify both records were inserted
	rows, err := db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to scan count: %v", err)
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}

func TestTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Insert initial record
	_, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "Initial", "initial@example.com")
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// Transaction that should fail and rollback
	err = Transaction(ctx, db, func(tx DBRunner) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
		if err != nil {
			return err
		}
		// This should fail due to unique constraint violation
		_, err = tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Jane", "initial@example.com")
		return err
	})

	if err == nil {
		t.Fatal("Transaction should have failed due to unique constraint violation")
	}

	// Verify only the initial record remains
	rows, err := db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to scan count: %v", err)
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 user (rollback should have occurred), got %d", count)
	}
}

func TestTransaction_RollbackOnError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testError := errors.New("test error")

	// Transaction that returns an error
	err := Transaction(ctx, db, func(tx DBRunner) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
		if err != nil {
			return err
		}
		return testError
	})

	if err == nil || err != testError {
		t.Fatalf("Expected test error, got: %v", err)
	}

	// Verify no records were inserted (rollback occurred)
	rows, err := db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to scan count: %v", err)
		}
	}

	if count != 0 {
		t.Errorf("Expected 0 users (rollback should have occurred), got %d", count)
	}
}

func TestTransaction_RollbackOnPanic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Test panic recovery and rollback
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic but none occurred")
		}
	}()

	Transaction(ctx, db, func(tx DBRunner) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
		if err != nil {
			return err
		}
		panic("test panic")
	})

	// This code should not be reached due to panic,
	// but if it is, we can verify rollback occurred
	rows, err := db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to scan count: %v", err)
		}
	}

	if count != 0 {
		t.Errorf("Expected 0 users (rollback should have occurred), got %d", count)
	}
}

func TestTransaction_NestedTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Start a transaction and pass it to another Transaction call
	err := Transaction(ctx, db, func(tx1 DBRunner) error {
		_, err := tx1.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
		if err != nil {
			return err
		}

		// This should reuse the existing transaction
		return Transaction(ctx, tx1, func(tx2 DBRunner) error {
			_, err := tx2.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Jane", "jane@example.com")
			return err
		})
	})

	if err != nil {
		t.Fatalf("Nested transaction failed: %v", err)
	}

	// Verify both records were inserted
	rows, err := db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			t.Fatalf("Failed to scan count: %v", err)
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}

func TestTransaction_Query(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Insert initial data
	_, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "John", "john@example.com")
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	var foundName string
	err = Transaction(ctx, db, func(tx DBRunner) error {
		rows, err := tx.QueryContext(ctx, "SELECT name FROM users WHERE email = ?", "john@example.com")
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			return rows.Scan(&foundName)
		}
		return errors.New("no user found")
	})

	if err != nil {
		t.Fatalf("Transaction with query failed: %v", err)
	}

	if foundName != "John" {
		t.Errorf("Expected 'John', got '%s'", foundName)
	}
}