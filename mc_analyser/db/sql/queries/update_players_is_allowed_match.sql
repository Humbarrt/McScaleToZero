UPDATE "players"
SET "is_allowed" = ?
WHERE "id" IN (
        SELECT value
        FROM json_each(?)
    );