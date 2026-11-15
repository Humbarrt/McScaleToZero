SELECT
    p.id AS player_id,
    COALESCE(pay.cnt, 0)   AS total_payments_count,
    COALESCE(pay.total, 0) AS total_paid_amount,
    COALESCE(inv.total, 0) AS total_invoiced_amount
FROM players p
LEFT JOIN (
    SELECT player_id, COUNT(*) AS cnt, SUM(amount) AS total
    FROM local_invoice_service_payments
    GROUP BY player_id
) pay ON pay.player_id = p.id
LEFT JOIN (
    SELECT player_id, SUM(amount) AS total
    FROM local_invoice_service_invoice_items
    GROUP BY player_id
) inv ON inv.player_id = p.id;
