INSERT INTO "billed_players" (
        "bill_id",
        "player_id",
        "total_amount",
        "fixed_amount",
        "variable_amount",
        "playtime",
        "cpu_time"
    )
VALUES (?, ?, ?, ?, ?, ?, ?);