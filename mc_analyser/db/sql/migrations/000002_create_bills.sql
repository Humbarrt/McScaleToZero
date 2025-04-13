-- +goose Up
CREATE TABLE IF NOT EXISTS "bills" (
    "id" INTEGER PRIMARY KEY,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "billing_month" INTEGER NOT NULL,
    "total_cost" INTEGER NOT NULL,
    "fixed_cost" INTEGER NOT NULL,
    "variable_cost" INTEGER NOT NULL,
    "currency" TEXT NOT NULL DEFAULT 'USD' CHECK (length("currency") <= 128),
    "invoice_service_invoice_id" TEXT NOT NULL DEFAULT '' CHECK (length("invoice_service_invoice_id") <= 1024)
);