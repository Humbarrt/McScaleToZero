-- +goose Up
-- a bit weird, true. But I don't won't to requery aws cost explorer on boot, each request costs 0.01$.
CREATE TABLE IF NOT EXISTS "explored_costs" (
    "id" INTEGER PRIMARY KEY,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "explored_month" INTEGER NOT NULL UNIQUE
);