SELECT COALESCE(SUM("total_amount"), 0) AS "billed_total_cost"
FROM "bills"