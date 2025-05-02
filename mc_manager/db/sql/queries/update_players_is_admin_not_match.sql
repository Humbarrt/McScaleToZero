UPDATE "players"
SET "is_admin" = ?
WHERE "id" NOT IN (
        SELECT value
        FROM json_each(?)
    );