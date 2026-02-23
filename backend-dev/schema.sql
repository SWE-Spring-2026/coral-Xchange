CREATE TABLE holdings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id FOREIGN KEY INTEGER,
    ticker TEXT,
    quantity INTEGER,
    price REAL
);

CREATE TABLE account (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    cash_balance REAL
);