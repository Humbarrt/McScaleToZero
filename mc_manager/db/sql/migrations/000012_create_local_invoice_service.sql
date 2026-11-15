-- +goose Up
CREATE TABLE IF NOT EXISTS "local_invoice_service_config" (
    "id" INTEGER PRIMARY KEY,
    "invoice_currency" VARCHAR(3) NOT NULL DEFAULT 'USD',
    "paying_player_id" INTEGER,
    FOREIGN KEY ("paying_player_id") REFERENCES "players" ("id")
);
INSERT INTO "local_invoice_service_config" ("invoice_currency")
VALUES ('USD');
CREATE TABLE IF NOT EXISTS "local_invoice_service_payments" (
    "id" INTEGER PRIMARY KEY,
    "player_id" INTEGER NOT NULL,
    "amount" INTEGER NOT NULL,
    "description" VARCHAR(255) NOT NULL,
    "edited_at" DATETIME,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("player_id") REFERENCES "players" ("id")
);
CREATE TABLE IF NOT EXISTS "local_invoice_service_invoices" (
    "id" INTEGER PRIMARY KEY,
    "amount" INTEGER NOT NULL,
    "description" TEXT NOT NULL,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS "local_invoice_service_invoice_items" (
    "id" INTEGER PRIMARY KEY,
    "invoice_id" INTEGER NOT NULL,
    "player_id" INTEGER NOT NULL,
    "amount" INTEGER NOT NULL,
    FOREIGN KEY ("invoice_id") REFERENCES "local_invoice_service_invoices" ("id"),
    FOREIGN KEY ("player_id") REFERENCES "players" ("id")
);
