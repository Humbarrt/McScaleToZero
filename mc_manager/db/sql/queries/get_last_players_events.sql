SELECT p.id,
    p.timestamp,
    p.player_id,
    p.is_online
FROM player_events p
WHERE p.timestamp < :before_time
    AND NOT EXISTS (
        SELECT 1
        FROM player_events p2
        WHERE p2.player_id = p.player_id
            AND p2.timestamp < :before_time
            AND p2.timestamp > p.timestamp
    )
ORDER BY p.player_id ASC;