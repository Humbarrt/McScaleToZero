-- +goose Up
CREATE TABLE IF NOT EXISTS "splitwise_config" (
    "id" INTEGER PRIMARY KEY,
    "invoice_currency" VARCHAR(3) NOT NULL DEFAULT 'USD',
    "currencies" TEXT NOT NULL DEFAULT '[]',
    "users" TEXT NOT NULL DEFAULT '[]'
);
INSERT INTO "splitwise_config" ("invoice_currency")
VALUES ('USD');