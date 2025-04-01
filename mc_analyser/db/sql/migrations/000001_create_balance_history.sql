-- +goose Up
CREATE TABLE IF NOT EXISTS "balance_history" (
    "id" INTEGER PRIMARY KEY,
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "usd" REAL NOT NULL CHECK(typeof("usd") = "real")
);