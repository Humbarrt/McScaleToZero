package queries

import (
	"encoding/json"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
)

func parsePlayer(row scannable) (*model.Player, error) {
	var id int
	var mcUUIDRaw []byte
	var isAllowed bool
	var isAdmin bool
	var lastIPV4 *string
	var aliasesString string
	var splitwiseId *string
	err := row.Scan(&id, &mcUUIDRaw, &isAllowed, &isAdmin, &lastIPV4, &aliasesString, &splitwiseId)
	if err != nil {
		return nil, err
	}

	aliases := []string{}
	if aliasesString != "" {
		err = json.Unmarshal([]byte(aliasesString), &aliases)
		if err != nil {
			return nil, err
		}
	}

	if len(aliases) == 0 {
		logging.Error("Player has no aliases", "id", id)
	}

	// Convert mcUUID to UUID type
	mcUUID, err := model.NewUUIDFromBytes(mcUUIDRaw[:])
	if err != nil {
		return nil, err
	}

	return &model.Player{
		Id:          id,
		McUUID:      *mcUUID,
		IsAllowed:   isAllowed,
		IsAdmin:     isAdmin,
		LastIPV4:    lastIPV4,
		Aliases:     aliases,
		SplitwiseId: splitwiseId,
	}, nil
}

func GetPlayers(DB *db.DB) ([]*model.Player, error) {
	stmt, err := DB.GetQuery("get_players")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	players := []*model.Player{}
	for result.Next() {
		player, err := parsePlayer(result)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func GetPlayerById(DB *db.DB, playerId int) (*model.Player, error) {
	stmt, err := DB.GetQuery("get_player_by_id")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(playerId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parsePlayer(result)
}

func GetPlayerByUUID(DB *db.DB, playerUUID model.UUID) (*model.Player, error) {
	stmt, err := DB.GetQuery("get_player_by_uuid")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(playerUUID.Bytes())
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parsePlayer(result)
}

func GetPlayerByAlias(DB *db.DB, alias string) (*model.Player, error) {
	stmt, err := DB.GetQuery("get_player_by_alias")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(alias)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parsePlayer(result)
}

func AddPlayer(DB *db.DB, mcUUID model.UUID, isAllowed bool, isAdmin bool, lastIPV4 *string, aliases []string, splitwiseId *string) error {
	stmt, err := DB.GetQuery("add_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mcUUID.Bytes(), isAllowed, isAdmin, lastIPV4, string(aliasesString), splitwiseId)
	if err != nil {
		return err
	}

	return nil
}

func AddPlayerAlias(DB *db.DB, playerId int, alias string) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	newAliases := []string{alias}
	for _, a := range player.Aliases {
		if a != alias {
			newAliases = append(newAliases, a)
		}
	}
	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, player.IsAdmin, player.LastIPV4, newAliases, player.SplitwiseId)
}

func UpdatePlayerSplitwiseId(DB *db.DB, playerId int, splitwiseId *string) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, player.IsAdmin, player.LastIPV4, player.Aliases, splitwiseId)
}

func UpdatePlayerIsAllowed(DB *db.DB, playerId int, isAllowed bool) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, isAllowed, player.IsAdmin, player.LastIPV4, player.Aliases, player.SplitwiseId)
}

func UpdatePlayerIsAdmin(DB *db.DB, playerId int, isAdmin bool) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, isAdmin, player.LastIPV4, player.Aliases, player.SplitwiseId)
}

func UpdatePlayerLastIPV4(DB *db.DB, playerId int, lastIPV4 *string) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, player.IsAdmin, lastIPV4, player.Aliases, player.SplitwiseId)
}

func updatePlayer(DB *db.DB, playerId int, mcUUID model.UUID, isAllowed bool, isAdmin bool, lastIPV4 *string, aliases []string, splitwiseId *string) error {
	stmt, err := DB.GetQuery("update_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mcUUID.Bytes(), isAllowed, isAdmin, lastIPV4, string(aliasesString), splitwiseId, playerId)
	if err != nil {
		return err
	}

	return nil
}
