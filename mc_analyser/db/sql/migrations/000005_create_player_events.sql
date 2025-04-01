-- +goose Up
CREATE TABLE IF NOT EXISTS "player_events" (
    "id" INTEGER PRIMARY KEY,
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "player_id" INTEGER NOT NULL,
    "is_online" BOOLEAN NOT NULL,
    FOREIGN KEY("player_id") REFERENCES "players"("id")
);