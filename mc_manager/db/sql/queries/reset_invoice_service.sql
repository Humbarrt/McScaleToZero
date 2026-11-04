--after the invoice service has been changed, reset the shared variables in the database
-- set bills.invoice_service_invoice_id to NULL everywhere
UPDATE bills
SET invoice_service_invoice_id = NULL;
--set player.invoice_service_user_id to NULL everywhere
UPDATE players
SET invoice_service_user_id = NULL;