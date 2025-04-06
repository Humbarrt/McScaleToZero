UPDATE "config"
SET "offline_image" = ?,
    "offline_motd" = ?,
    "mc_port" = ?,
    "ip_check" = ?
WHERE "id" = 1;