-- +goose Up
CREATE TABLE IF NOT EXISTS "players" (
    "id" INTEGER PRIMARY KEY,
    "mc_uuid" BLOB NOT NULL UNIQUE,
    "is_allowed" BOOLEAN NOT NULL DEFAULT TRUE,
    "is_admin" BOOLEAN NOT NULL DEFAULT FALSE,
    "last_ipv4" TEXT CHECK (length("last_ipv4") <= 15),
    "aliases" TEXT NOT NULL DEFAULT '[]' CHECK (json_valid("aliases")),
    "invoice_service_user_id" TEXT,
    "client_key" BLOB CHECK(
        "client_key" IS NULL
        OR (
            typeof("client_key") = "blob"
            AND length("client_key") <= 2048
        )
    ),
    "encrypted_password" BLOB CHECK(
        "encrypted_password" IS NULL
        OR (
            typeof("encrypted_password") = "blob"
            AND length("encrypted_password") <= 256
        )
    )
);