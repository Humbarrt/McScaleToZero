SELECT "id",
    "mc_id",
    "is_allowed",
    "aliases",
    "splitwise_id"
FROM "players"
WHERE EXISTS (
        SELECT 1
        FROM json_each("aliases")
        WHERE value = ?
    );