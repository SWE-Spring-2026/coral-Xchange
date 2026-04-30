package main

import (
	"database/sql"
	"fmt"
	"log"
)

// ---------------------------------------------------------------------------
// Database initialisation
// ---------------------------------------------------------------------------

func initDatabase(database *sql.DB) {
	// Enable foreign-key enforcement (off by default in SQLite).
	_, err := database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	queries := []string{
		// Users table — the anchor for every other table.
		`CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			name          TEXT NOT NULL,
			username      TEXT UNIQUE NOT NULL,
			email         TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// One account row per user, created automatically on registration.
		`CREATE TABLE IF NOT EXISTS account (
			user_id      INTEGER PRIMARY KEY,
			cash_balance REAL NOT NULL DEFAULT 100000.00,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Each row is one purchase lot for one user.
		// Multiple rows per (user_id, ticker) are intentional — each BUY
		// creates a new lot so the purchase price is preserved for
		// realized P&L calculation. purchased_at drives FIFO ordering
		// when shares are sold.
		`CREATE TABLE IF NOT EXISTS holdings (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER  NOT NULL,
			ticker       TEXT     NOT NULL,
			quantity     INTEGER  NOT NULL,
			price        REAL     NOT NULL,
			purchased_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Each row is one completed SELL trade.
		// cost_basis   — what the sold shares originally cost (sum across lots consumed)
		// realized_pnl — profit (positive) or loss (negative) locked in by the sale
		`CREATE TABLE IF NOT EXISTS trades (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER NOT NULL,
			ticker       TEXT    NOT NULL,
			quantity     INTEGER NOT NULL,
			sell_price   REAL    NOT NULL,
			cost_basis   REAL    NOT NULL,
			realized_pnl REAL    NOT NULL,
			sold_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := database.Exec(q); err != nil {
			log.Fatal("Failed to run init query:", err)
		}
	}

	fmt.Println("Database initialised successfully.")
}