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
	defer func() {
		_ = db.Close()
	}()

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
