WITH last_event_before AS (
    SELECT *
    FROM server_events
    WHERE timestamp < :start_time
    ORDER BY timestamp DESC
    LIMIT 1
), state_before AS (
    SELECT :start_time AS timestamp,
        1 AS is_online
    FROM last_event_before
    WHERE is_online = 1
),
events_in_range AS (
    SELECT timestamp,
        is_online
    FROM server_events
    WHERE timestamp BETWEEN :start_time AND :end_time
),
combined_events AS (
    SELECT *
    FROM events_in_range
    UNION ALL
    SELECT *
    FROM state_before
),
ordered_events AS (
    SELECT timestamp,
        is_online,
        LEAD(timestamp) OVER (
            ORDER BY timestamp
        ) AS next_timestamp
    FROM combined_events
),
sessions AS (
    SELECT timestamp AS session_start,
        COALESCE(next_timestamp, :end_time) AS session_end
    FROM ordered_events
    WHERE is_online = 1
),
valid_sessions AS (
    SELECT session_start,
        session_end
    FROM sessions
    WHERE session_start < session_end
)
SELECT COALESCE(
        SUM(
            strftime('%s', session_end) - strftime('%s', session_start)
        ),
        0
    ) AS total_seconds_online
FROM valid_sessions;