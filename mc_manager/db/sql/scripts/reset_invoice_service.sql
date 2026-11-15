--after the invoice service has been changed, reset the shared variables in the database
-- set bills.invoice_service_invoice_id to NULL everywhere
UPDATE bills
SET invoice_service_invoice_id = NULL;

--set player.invoice_service_user_id to NULL everywhere
UPDATE players
SET invoice_service_user_id = NULL;

--Reset Splitwise Tables
DELETE FROM splitwise_config;

--Reset Local Invoice Service Tables
UPDATE "local_invoice_service_config" SET paying_player_id = NULL;
DELETE FROM local_invoice_service_invoice_items;
DELETE FROM local_invoice_service_invoices;
DELETE FROM local_invoice_service_payments;
