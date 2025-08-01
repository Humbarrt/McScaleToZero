INSERT INTO last_log (id, timestamp, message)
VALUES (1, ?, ?) ON CONFLICT(id) DO
UPDATE
SET timestamp = excluded.timestamp,
    message = excluded.message;