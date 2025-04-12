-- +goose Up
CREATE TABLE IF NOT EXISTS "config" (
    "id" INTEGER PRIMARY KEY,
    "offline_image" BLOB,
    "offline_motd" TEXT CHECK (
        "offline_motd" IS NULL
        OR (
            typeof("offline_motd") = 'text'
            AND length("offline_motd") <= 1024
        )
    ),
    "mc_port" INTEGER NOT NULL CHECK (
        typeof("mc_port") = 'integer'
        AND "mc_port" > 0
        AND "mc_port" <= 65535
    ),
    "ip_check" BOOLEAN NOT NULL,
    "backups_to_keep" INTEGER NOT NULL CHECK (
        typeof("backups_to_keep") = 'integer'
        AND "backups_to_keep" >= 0
        AND "backups_to_keep" <= 15
    )
);
INSERT INTO "config" (
        "offline_image",
        "offline_motd",
        "mc_port",
        "ip_check",
        "backups_to_keep"
    )
VALUES (NULL, NULL, 25565, TRUE, 2);