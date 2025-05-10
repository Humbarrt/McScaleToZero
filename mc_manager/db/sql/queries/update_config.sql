UPDATE "config"
SET "offline_image" = ?,
    "offline_motd" = ?,
    "mc_port" = ?,
    "ip_check" = ?,
    "backups_to_keep" = ?,
    "tax_rate" = ?
WHERE "id" = 1;