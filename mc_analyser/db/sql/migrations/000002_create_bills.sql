-- +goose Up
CREATE TABLE IF NOT EXISTS "bills" (
    "id" INTEGER PRIMARY KEY,
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "usd" REAL NOT NULL CHECK(typeof("usd") = "real"),
    "eur" REAL NOT NULL CHECK(typeof("eur") = "real"),
    "old_balance" REAL NOT NULL CHECK(typeof("old_balance") = "real"),
    "new_balance" REAL NOT NULL CHECK(typeof("new_balance") = "real")
);