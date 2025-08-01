-- +goose Up
CREATE TABLE IF NOT EXISTS "server_events" (
    "id" INTEGER PRIMARY KEY,
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "is_online" BOOLEAN NOT NULL
);
CREATE INDEX IF NOT EXISTS "idx_server_events_timestamp" ON "server_events"("timestamp");