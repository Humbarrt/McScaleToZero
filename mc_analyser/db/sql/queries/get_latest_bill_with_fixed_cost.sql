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
WHERE "fixed_amount" IS NOT NULL
    AND "fixed_amount" > 0
ORDER BY "billing_month" DESC
LIMIT 1;