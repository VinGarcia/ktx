# ktx - Simple & Safe Database Transactions

A Keep it simple Transaction manager that simplifies error handling and provides
safe database transaction management with automatic rollback when errors or panics occur.

## Features

- **Panic-safe**: Automatically rolls back transactions if a panic occurs
- **Error handling**: Rolls back transactions if an error is returned
- **Nested transaction support**: Reuses existing transactions when called within another transaction
- **Compatible with database/sql**: Works with all databases supported by `database/sql`

## Usage

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vingarcia/ktx"
)

func main() {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createUsersTable(db)

	err = ktx.Transaction(ctx, db, func(db *sql.Tx) error {
		_, err := db.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?,  ?)", "John", "john@gmail.com")
		if err != nil {
			return err
		}

		var name string
		err = db.QueryRow("SELECT name FROM users WHERE email = ?", "john@gmail.com").Scan(&name)
		if err != nil {
			return err
		}

		fmt.Printf("successfully inserted user %s inside transaction\n", name)

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func createUsersTable(db *sql.DB) {
	_, _ = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)
	`)
}
```

## Lint & Testing

Run the lint and tests with:

```bash
make lint test
```

The tests use an in-memory SQLite database and cover:
- Successful transactions
- Rollback on database errors
- Rollback on explicit errors
- Rollback on panics
- Nested transaction handling
- Query operations within transactions
