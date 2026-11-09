package server_watcher

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"mc_manager/api/ssh"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"slices"
	"time"
)

type mcPlayer struct {
	UUID model.UUID `json:"uuid"`
	Name string     `json:"name"`
}

func updatePlayersInDB(db *db.DB, players []*mcPlayer) (*[]int64, error) {
	playerIds := []int64{}
	for _, mcPlayer := range players {
		dbPlayer, err := queries.GetPlayerByUUID(db, mcPlayer.UUID)
		if err != nil && err != sql.ErrNoRows {
			return &playerIds, err
		}
		if err == sql.ErrNoRows || dbPlayer == nil {
			err := queries.AddPlayer(db, mcPlayer.UUID, true, false, nil, []string{mcPlayer.Name}, nil)
			if err != nil {
				return &playerIds, err
			}
			dbPlayer, err = queries.GetPlayerByUUID(db, mcPlayer.UUID)
			if err != nil {
				return &playerIds, err
			}
		}

		err = queries.AddPlayerAlias(db, dbPlayer.Id, mcPlayer.Name)
		if err != nil {
			return &playerIds, err
		}
		playerIds = append(playerIds, dbPlayer.Id)
	}

	if addedNewPlayer {
		go func() {
			time.Sleep(100 * time.Millisecond)
			app.EB.Broadcast(model.EB_EVENT_NEW_PLAYER)
		}()
	}

	return &playerIds, nil
}

func parseMcPlayerFile(data *bufio.Reader) ([]*mcPlayer, error) {
	playerFile := make([]mcPlayer, 0)
	err := json.NewDecoder(data).Decode(&playerFile)
	if err != nil {
		return nil, err
	}
	logging.Debug("PlayerFile loaded", "players", playerFile)

	playersPointers := make([]*mcPlayer, len(playerFile))
	for i := range playerFile {
		playersPointers[i] = &playerFile[i]
	}

	return playersPointers, nil
}

func downloadMcFiles(app *model.App) ([]*mcPlayer, []*mcPlayer, error) {
	sshClient, err := ssh.NewClient(app)
	if err != nil {
		return nil, nil, err
	}
	defer sshClient.Close()

	// Read the whitelist file from the SFTP server
	whitelistReader, err := sshClient.ReadFile("./mc-server/data/whitelist.json")
	if err != nil {
		return nil, nil, err
	}
	whitelist, err := parseMcPlayerFile(whitelistReader)
	if err != nil {
		return nil, nil, err
	}

	// Read the admin file from the SFTP server
	adminReader, err := sshClient.ReadFile("./mc-server/data/ops.json")
	if err != nil {
		return nil, nil, err
	}
	admins, err := parseMcPlayerFile(adminReader)
	return whitelist, admins, err
}

func processMcPlayers(db *db.DB, whitelist []*mcPlayer, admins []*mcPlayer) error {
	whitelistPlayerIds, err := updatePlayersInDB(db, whitelist)
	if err != nil {
		return err
	}
	adminPlayerIds, err := updatePlayersInDB(db, admins)
	if err != nil {
		return err
	}

	err = queries.UpdatePlayersIsAdmin(db, *adminPlayerIds, true, true)
	if err != nil {
		return err
	}
	err = queries.UpdatePlayersIsAdmin(db, *adminPlayerIds, false, false)
	if err != nil {
		return err
	}

	//merge whitelist and admin player ids
	mergedPlayerIds := append(*whitelistPlayerIds, *adminPlayerIds...)
	slices.Sort(mergedPlayerIds)
	mergedPlayerIds = slices.Compact(mergedPlayerIds)

	err = queries.UpdatePlayersIsAllowed(db, mergedPlayerIds, true, true)
	if err != nil {
		return err
	}
	return queries.UpdatePlayersIsAllowed(db, mergedPlayerIds, false, false)
}

func RemoveSessionsFromNonWhitelistedPlayers(app *model.App) error {
	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error getting players from database", "error", err)
		return err
	}

	for _, player := range players {
		if player.IsAllowed || player.IsAdmin {
			continue
		}
		err := app.SM.RemoveUserSessions(player.Id)
		if err != nil {
			logging.Error("Error removing sessions for player", "playerId", player.Id, "error", err)
		}
	}
	return nil
}

func UpdateWhitelistAndAdmins(app *model.App) error {
	whitelist, admins, err := downloadMcFiles(app)
	if err != nil {
		logging.Warn("Error downloading player files, trying again.", "error", err)

		time.Sleep(5 * time.Second)
		whitelist, admins, err = downloadMcFiles(app)
		if err != nil {
			logging.Error("Error downloading player files, giving up.", "error", err)
			return err
		}
	}

	logging.Debug("Whitelist and Admins loaded", "whitelist", whitelist, "admins", admins)
	err = processMcPlayers(app.Db, whitelist, admins)
	if err != nil {
		logging.Error("Error processing players", "error", err)
		return err
	}

	return RemoveSessionsFromNonWhitelistedPlayers(app)
}
