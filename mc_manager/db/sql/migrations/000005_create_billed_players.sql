-- +goose Up
CREATE TABLE IF NOT EXISTS "billed_players" (
    "id" INTEGER PRIMARY KEY,
    "bill_id" INTEGER NOT NULL,
    "player_id" INTEGER NOT NULL,
    "total_amount" INTEGER NOT NULL CHECK(
        typeof("total_amount") = "integer"
        AND "total_amount" >= 0
    ),
    "fixed_amount" INTEGER NOT NULL CHECK(
        typeof("fixed_amount") = "integer"
        AND "fixed_amount" >= 0
    ),
    "variable_amount" INTEGER NOT NULL CHECK(
        typeof("variable_amount") = "integer"
        AND "variable_amount" >= 0
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
CREATE INDEX IF NOT EXISTS "idx_billed_players_bill_id_player_id" ON "billed_players"("bill_id");