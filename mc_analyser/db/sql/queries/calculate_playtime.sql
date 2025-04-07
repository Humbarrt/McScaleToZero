WITH last_event_before AS (
    SELECT player_id,
        MAX(timestamp) AS last_time
    FROM player_events
    WHERE timestamp < ?
    GROUP BY player_id
),
state_before AS (
    SELECT e.player_id,
        e.is_online
    FROM player_events e
        JOIN last_event_before l ON e.player_id = l.player_id
        AND e.timestamp = l.last_time
),
events_in_range AS (
    SELECT player_id,
        timestamp,
        is_online,
        LEAD(timestamp) OVER (
            PARTITION BY player_id
            ORDER BY timestamp
        ) AS next_timestamp
    FROM player_events
    WHERE timestamp BETWEEN ? AND ?
),
synthetic_start_events AS (
    SELECT sb.player_id,
        ? AS timestamp,
        -- start_date
        1 AS is_online,
        MIN(e.timestamp) AS next_timestamp
    FROM state_before sb
        LEFT JOIN events_in_range e ON sb.player_id = e.player_id
        AND e.is_online = 0
    WHERE sb.is_online = 1
    GROUP BY sb.player_id
),
combined_events AS (
    SELECT *
    FROM events_in_range
    UNION ALL
    SELECT *
    FROM synthetic_start_events
),
sessions AS (
    SELECT player_id,
        MAX(timestamp, ?) AS session_start,
        MIN(COALESCE(next_timestamp, ?), ?) AS session_end
    FROM combined_events
    WHERE is_online = 1
),
valid_sessions AS (
    SELECT player_id,
        session_start,
        session_end
    FROM sessions
    WHERE session_start < session_end
)
SELECT player_id,
    SUM(
        strftime('%s', session_end) - strftime('%s', session_start)
    ) AS total_seconds_online
FROM valid_sessions
GROUP BY player_id;