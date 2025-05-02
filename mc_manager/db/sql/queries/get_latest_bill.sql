SELECT "id",
    "created_at",
    "billing_month",
    "total_amount",
    "fixed_amount",
    "variable_amount",
    "server_uptime",
    "currency",
    "invoice_service_invoice_id"
FROM "bills"
ORDER BY "billing_month" DESC
LIMIT 1;