-- +goose Up
CREATE TABLE IF NOT EXISTS "bill_items" (
    "id" INTEGER PRIMARY KEY,
    "bill_id" INTEGER NOT NULL,
    "label" TEXT NOT NULL CHECK (length("label") <= 255),
    "is_fixed" BOOLEAN NOT NULL DEFAULT FALSE,
    "cost" INTEGER NOT NULL CHECK(
        typeof("cost") = "integer"
        AND "cost" >= 0
    ),
    FOREIGN KEY("bill_id") REFERENCES "bills"("id") ON DELETE CASCADE
);