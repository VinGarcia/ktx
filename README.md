# ktx - Safe Database Transactions

A simple Go package that provides safe database transaction management with automatic rollback on errors and panics.

## Features

- **Panic-safe**: Automatically rolls back transactions if a panic occurs
- **Error handling**: Rolls back transactions on any error
- **Nested transaction support**: Reuses existing transactions when called within another transaction
- **Compatible with database/sql**: Works with any database driver that implements the standard interfaces

## Usage

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    "github.com/vingarcia/ktx"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "example.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    err = ktx.Transaction(ctx, db, func(tx ktx.DBRunner) error {
        _, err := tx.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "John")
        if err != nil {
            return err
        }
        
        _, err = tx.ExecContext(ctx, "INSERT INTO posts (title, author) VALUES (?, ?)", "Hello World", "John")
        return err
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

## Interface Compatibility

The package defines minimal interfaces that are compatible with the standard `database/sql` package:

- `DBRunner`: Compatible with `*sql.DB` and `*sql.Tx`
- `TxBeginner`: Compatible with `*sql.DB`
- `Tx`: Compatible with `*sql.Tx`

## Testing

Run the tests with:

```bash
go test
```

The tests use an in-memory SQLite database and cover:
- Successful transactions
- Rollback on database errors
- Rollback on explicit errors
- Rollback on panics
- Nested transaction handling
- Query operations within transactions