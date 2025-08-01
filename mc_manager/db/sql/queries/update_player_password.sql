UPDATE "players"
SET "encrypted_password" = ?,
    "client_key" = ?
WHERE "id" = ?;