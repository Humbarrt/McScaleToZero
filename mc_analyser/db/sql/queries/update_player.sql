UPDATE "players"
SET "mc_uuid" = ?,
    "is_allowed" = ?,
    "aliases" = ?,
    "splitwise_id" = ?
WHERE "id" = ?;