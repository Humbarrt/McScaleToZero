package queries

import (
	"encoding/json"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
)

func parsePlayer(row scannable) (*model.Player, error) {
	var id int64
	var mcUUIDRaw []byte
	var isAllowed bool
	var aliasesString string
	var splitwiseId string
	err := row.Scan(&id, &mcUUIDRaw, &isAllowed, &aliasesString, &splitwiseId)
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
	mcUUID, err := model.UUIDFromBytes(mcUUIDRaw[:])
	if err != nil {
		return nil, err
	}

	return &model.Player{
		Id:          id,
		McUUID:      *mcUUID,
		IsAllowed:   isAllowed,
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

func GetPlayerByUUID(DB *db.DB, playerUUID int) (*model.Player, error) {
	stmt, err := DB.GetQuery("get_player_by_uuid")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(playerUUID)
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

func AddPlayer(DB *db.DB, mcUUID model.UUID, isAllowed bool, aliases []string, splitwiseId string) error {
	stmt, err := DB.GetQuery("add_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mcUUID.Bytes(), isAllowed, string(aliasesString), splitwiseId)
	if err != nil {
		return err
	}

	return nil
}

func AddPlayerAlias(DB *db.DB, playerId int64, alias string) error {
	player, err := GetPlayerById(DB, int(playerId))
	if err != nil {
		return err
	}

	newAliases := []string{alias}
	for _, a := range player.Aliases {
		if a != alias {
			newAliases = append(newAliases, a)
		}
	}
	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, newAliases, player.SplitwiseId)
}

func SetPlayerSplitwiseId(DB *db.DB, playerId int64, splitwiseId string) error {
	player, err := GetPlayerById(DB, int(playerId))
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, player.IsAllowed, player.Aliases, splitwiseId)
}

func SetPlayerIsAllowed(DB *db.DB, playerId int64, isAllowed bool) error {
	player, err := GetPlayerById(DB, int(playerId))
	if err != nil {
		return err
	}

	return updatePlayer(DB, playerId, player.McUUID, isAllowed, player.Aliases, player.SplitwiseId)
}

func updatePlayer(DB *db.DB, playerId int64, mcUUID model.UUID, isAllowed bool, aliases []string, splitwiseId string) error {
	stmt, err := DB.GetQuery("update_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mcUUID.Bytes(), isAllowed, string(aliasesString), splitwiseId, playerId)
	if err != nil {
		return err
	}

	return nil
}
