SELECT "id",
    "bill_id",
    "label",
    "is_fixed",
    "amount"
FROM "bill_items"
WHERE "bill_id" = $1
ORDER BY "id" ASC