UPDATE "players"
SET "mc_id" = ?,
    "is_allowed" = ?,
    "aliases" = ?,
    "splitwise_id" = ?
WHERE "id" = ?;