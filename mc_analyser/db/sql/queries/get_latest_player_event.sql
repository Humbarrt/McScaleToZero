SELECT "id",
    "timestamp",
    "player_id",
    "is_online"
FROM "player_events"
WHERE "player_id" = ?
ORDER BY "timestamp" DESC
LIMIT 1;