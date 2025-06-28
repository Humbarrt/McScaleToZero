-- +goose Up
ALTER TABLE "players" ADD COLUMN "avatar_head" BLOB;
ALTER TABLE "players" ADD COLUMN "last_avatar_head_download" DATETIME;