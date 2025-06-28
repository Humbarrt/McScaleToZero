SELECT "player_id",
    COALESCE(SUM("playtime"), 0),
    COALESCE(SUM("cpu_time"), 0)
FROM "billed_players"
GROUP BY "player_id"