WITH players_with_alias AS (
    SELECT p.id
    FROM players p,
        json_each(p.aliases)
    WHERE json_each.value = ?
        AND p.id != ?
)
UPDATE players
SET aliases = (
        SELECT json_group_array(value)
        FROM json_each(players.aliases)
        WHERE value != ?
    )
WHERE players.id IN (
        SELECT id
        FROM players_with_alias
    );