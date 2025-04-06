-- +goose Up
CREATE TABLE IF NOT EXISTS "config" (
    "id" INTEGER PRIMARY KEY,
    "offline_image" BLOB,
    "offline_motd" TEXT,
    "mc_port" INTEGER NOT NULL,
    "ip_check" BOOLEAN NOT NULL,
    CHECK (
        offline_motd IS NULL
        OR (
            typeof(offline_motd) = 'text'
            AND length(offline_motd) <= 1024
        )
    )
);
INSERT INTO "config" (
        "offline_image",
        "offline_motd",
        "mc_port",
        "ip_check"
    )
VALUES (NULL, NULL, 25565, TRUE);