SELECT "id",
    "bill_id",
    "player_id",
    "total_amount",
    "fixed_amount",
    "variable_amount",
    "playtime",
    "cpu_time"
FROM "billed_players"
WHERE "bill_id" = $1
ORDER BY "id" ASC;