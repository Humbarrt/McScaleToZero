SELECT "id",
    "mc_uuid",
    "is_allowed",
    "is_admin",
    "last_ipv4",
    "aliases",
    "splitwise_id"
FROM "players"
WHERE EXISTS (
        SELECT 1
        FROM json_each("aliases")
        WHERE value = ?
    );