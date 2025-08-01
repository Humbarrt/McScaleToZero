SELECT "id",
    "timestamp",
    "player_id",
    "is_online"
FROM "player_events"
WHERE "timestamp" BETWEEN :start_time AND :end_time
ORDER BY "timestamp" ASC;