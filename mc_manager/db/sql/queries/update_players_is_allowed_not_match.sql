UPDATE "players"
SET "is_allowed" = ?
WHERE "id" NOT IN (
        SELECT value
        FROM json_each(?)
    );