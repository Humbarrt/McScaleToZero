UPDATE "players"
SET "is_admin" = ?
WHERE "id" IN (
        SELECT value
        FROM json_each(?)
    );