-- +goose Up
ALTER TABLE "config" ADD COLUMN "last_invoice_service" TEXT;
