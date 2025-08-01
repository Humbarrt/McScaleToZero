-- +goose Up
CREATE TABLE IF NOT EXISTS "bill_items" (
    "id" INTEGER PRIMARY KEY,
    "bill_id" INTEGER NOT NULL,
    "label" TEXT NOT NULL CHECK (length("label") <= 255),
    "is_fixed" BOOLEAN NOT NULL DEFAULT FALSE,
    "amount" INTEGER NOT NULL CHECK(
        typeof("amount") = "integer"
        AND "amount" >= 0
    ),
    FOREIGN KEY("bill_id") REFERENCES "bills"("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_bill_items_bill_id_label" ON "bill_items"("bill_id");