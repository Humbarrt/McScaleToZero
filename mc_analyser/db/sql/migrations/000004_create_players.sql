-- +goose Up
CREATE TABLE IF NOT EXISTS "players" (
    "id" INTEGER PRIMARY KEY,
    "mc_uuid" BLOB NOT NULL UNIQUE,
    "is_allowed" BOOLEAN NOT NULL DEFAULT TRUE,
    "is_admin" BOOLEAN NOT NULL DEFAULT FALSE,
    "last_ipv4" TEXT CHECK (length("last_ipv4") <= 15),
    "aliases" TEXT NOT NULL DEFAULT '[]' CHECK (json_valid("aliases")),
    "splitwise_id" TEXT
);