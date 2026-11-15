SELECT "id", "player_id", "amount", "description", "edited_at", "created_at" FROM "local_invoice_service_payments" WHERE "player_id" = ? ORDER BY COALESCE("edited_at", "created_at") DESC;
