UPDATE "players"
SET "mc_uuid" = ?,
    "is_allowed" = ?,
    "is_admin" = ?,
    "last_ipv4" = ?,
    "aliases" = ?,
    "invoice_service_user_id" = ?
WHERE "id" = ?;