-- +goose Up
CREATE TABLE IF NOT EXISTS "config" (
    "id" INTEGER PRIMARY KEY,
    "mc_port" INTEGER NOT NULL CHECK (
        typeof("mc_port") = 'integer'
        AND "mc_port" > 0
        AND "mc_port" <= 65535
    ),
    "online_image" BLOB,
    "online_motd" BLOB NOT NULL CHECK (length("online_motd") <= 1024),
    "offline_image" BLOB,
    "offline_motd" BLOB NOT NULL CHECK (length("offline_motd") <= 1024),
    "replace_motd" BOOLEAN NOT NULL,
    "last_image" BLOB,
    "last_motd" BLOB NOT NULL CHECK (length("last_motd") <= 1024),
    "protocol_version" BLOB NOT NULL CHECK (length("protocol_version") > 0),
    "max_players" INTEGER NOT NULL CHECK (
        typeof("max_players") = 'integer'
        AND "max_players" >= 0
        AND "max_players" <= 10000
    ),
    "backups_to_keep" INTEGER NOT NULL CHECK (
        typeof("backups_to_keep") = 'integer'
        AND "backups_to_keep" >= 0
        AND "backups_to_keep" <= 15
    ),
    "tax_rate" INTEGER NOT NULL CHECK (
        typeof("tax_rate") = 'integer'
        AND "tax_rate" >= 0
        AND "tax_rate" <= 10000
    ),
    "network_ratio" INTEGER NOT NULL CHECK (
        typeof("network_ratio") = 'integer'
        AND "network_ratio" >= 0
        AND "network_ratio" <= 10000
    ),
    "fallback_fixed_cost" INTEGER NOT NULL CHECK (
        typeof("fallback_fixed_cost") = 'integer'
        AND "fallback_fixed_cost" >= 0
    ),
    "fallback_variable_cost" INTEGER NOT NULL CHECK (
        typeof("fallback_variable_cost") = 'integer'
        AND "fallback_variable_cost" >= 0
    )
);
INSERT INTO "config" (
        "online_image",
        "online_motd",
        "offline_image",
        "offline_motd",
        "replace_motd",
        "last_image",
        "last_motd",
        "protocol_version",
        "max_players",
        "mc_port",
        "backups_to_keep",
        "tax_rate",
        "network_ratio",
        "fallback_fixed_cost",
        "fallback_variable_cost"
    )
VALUES (
        NULL,
        '\n                    \u00A7a(running)',
        NULL,
        '\n                \u00A78(not started yet)',
        FALSE,
        NULL,
        'Start your server to replace this message.',
        '{"name": "1.21.5", "protocol": 770}',
        20,
        25565,
        2,
        1900,
        1300,
        500,
        500
    );