SELECT "id",
    "timestamp",
    "player_id",
    "is_online"
FROM "player_events"
WHERE "timestamp" = (
        SELECT MAX("timestamp")
        FROM "player_events" AS pe
        WHERE pe."player_id" = player_events."player_id"
    )
ORDER BY "player_id" ASC;