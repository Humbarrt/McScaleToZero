SELECT ii."id", ii."amount", i."description", i."created_at" FROM "local_invoice_service_invoice_items" ii
LEFT JOIN "local_invoice_service_invoices" i ON ii."invoice_id" = i."id"
WHERE ii."player_id" = ?;
