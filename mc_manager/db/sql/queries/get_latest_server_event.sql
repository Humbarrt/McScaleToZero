SELECT "id",
    "timestamp",
    "is_online"
FROM "server_events"
ORDER BY "timestamp" DESC
LIMIT 1;