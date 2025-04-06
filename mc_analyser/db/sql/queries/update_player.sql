UPDATE "players"
SET "mc_uuid" = ?,
    "is_allowed" = ?,
    "is_admin" = ?,
    "last_ipv4" = ?,
    "aliases" = ?,
    "splitwise_id" = ?
WHERE "id" = ?;