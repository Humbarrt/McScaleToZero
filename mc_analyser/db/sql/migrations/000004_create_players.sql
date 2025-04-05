-- +goose Up
CREATE TABLE IF NOT EXISTS "players" (
    "id" INTEGER PRIMARY KEY,
    "mc_uuid" BLOB NOT NULL UNIQUE,
    "is_allowed" BOOLEAN NOT NULL DEFAULT TRUE,
    "aliases" TEXT CHECK (json_valid("aliases")),
    "splitwise_id" TEXT
);