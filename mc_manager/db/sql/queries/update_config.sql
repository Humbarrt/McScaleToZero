UPDATE "config"
SET "mc_port" = ?,
    "online_image" = ?,
    "online_motd" = ?,
    "offline_image" = ?,
    "offline_motd" = ?,
    "replace_motd" = ?,
    "last_image" = ?,
    "last_motd" = ?,
    "protocol_version" = ?,
    "max_players" = ?,
    "backups_to_keep" = ?,
    "tax_rate" = ?,
    "network_ratio" = ?,
    "fallback_fixed_cost" = ?,
    "fallback_variable_cost" = ?
WHERE "id" = 1;