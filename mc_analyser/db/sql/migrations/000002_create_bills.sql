-- +goose Up
CREATE TABLE IF NOT EXISTS "bills" (
    "id" INTEGER PRIMARY KEY,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "billing_month" INTEGER NOT NULL UNIQUE,
    "total_amount" INTEGER NOT NULL CHECK(
        typeof("total_amount") = "integer"
        AND "total_amount" >= 0
    ),
    "fixed_amount" INTEGER NOT NULL CHECK(
        typeof("fixed_amount") = "integer"
        AND "fixed_amount" >= 0
    ),
    "variable_amount" INTEGER NOT NULL CHECK(
        typeof("variable_amount") = "integer"
        AND "variable_amount" >= 0
    ),
    "server_uptime" INTEGER NOT NULL CHECK(
        typeof("server_uptime") = "integer"
        AND "server_uptime" >= 0
    ),
    "currency" TEXT NOT NULL DEFAULT 'USD' CHECK (length("currency") <= 128),
    "invoice_service_invoice_id" TEXT CHECK (length("invoice_service_invoice_id") <= 1024)
);