SELECT "billed_players"."player_id",
    COALESCE(SUM("billed_players"."playtime"), 0),
    COALESCE(SUM("billed_players"."cpu_time"), 0)
FROM "billed_players"
    RIGHT JOIN "bills" ON "bills"."id" = "billed_players"."bill_id"
WHERE "bills"."billing_month" >= ?
    AND "bills"."billing_month" < ?
GROUP BY "billed_players"."player_id"