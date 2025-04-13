-- +goose Up
CREATE TABLE IF NOT EXISTS "billed_players" (
    "id" INTEGER PRIMARY KEY,
    "bill_id" INTEGER NOT NULL,
    "player_id" INTEGER NOT NULL,
    "cost" INTEGER NOT NULL CHECK(
        typeof("cost") = "integer"
        AND "cost" >= 0
    ),
    "playtime" INTEGER NOT NULL CHECK(
        typeof("playtime") = "integer"
        AND "playtime" >= 0
    ),
    "cpu_time" INTEGER NOT NULL CHECK(
        typeof("cpu_time") = "integer"
        AND "cpu_time" >= 0
    ),
    FOREIGN KEY("bill_id") REFERENCES "bills"("id") ON DELETE CASCADE,
    FOREIGN KEY("player_id") REFERENCES "players"("id") ON DELETE CASCADE
);