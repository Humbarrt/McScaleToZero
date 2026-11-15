SELECT "id", "player_id", "amount", "description", "edited_at", "created_at" FROM "local_invoice_service_payments" ORDER BY COALESCE("edited_at", "created_at") DESC;
