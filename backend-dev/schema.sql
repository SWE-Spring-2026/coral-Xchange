-- coral-xchange schema
-- Run "PRAGMA foreign_keys = ON" at connection time (handled in Go).
 
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT UNIQUE NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);
 
-- One account per user, provisioned automatically on registration.
CREATE TABLE IF NOT EXISTS account (
    user_id      INTEGER PRIMARY KEY,
    cash_balance REAL NOT NULL DEFAULT 100000.00,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
 
-- Each row is one ticker position for one user.
CREATE TABLE IF NOT EXISTS holdings (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id  INTEGER NOT NULL,
    ticker   TEXT    NOT NULL,
    quantity INTEGER NOT NULL,
    price    REAL    NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
 