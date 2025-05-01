package queries

import (
	"encoding/json"
	"fmt"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
)

func parsePlayer(row scannable) (*model.Player, error) {
	var id int64
	var mcUUIDRaw []byte
	var isAllowed bool
	var isAdmin bool
	var lastIPV4 *string
	var aliasesString string
	var invoiceServiceUserId *string
	err := row.Scan(&id, &mcUUIDRaw, &isAllowed, &isAdmin, &lastIPV4, &aliasesString, &invoiceServiceUserId)
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
		Id:                   id,
		McUUID:               *mcUUID,
		IsAllowed:            isAllowed,
		IsAdmin:              isAdmin,
		LastIPV4:             lastIPV4,
		Aliases:              aliases,
		InvoiceServiceUserId: invoiceServiceUserId,
	}, nil
}

func GetPlayers(DB *db.DB) ([]*model.Player, error) {
	stmt, err := DB.GetStmt("get_players")
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

func GetPlayerById(DB *db.DB, playerId int64) (*model.Player, error) {
	stmt, err := DB.GetStmt("get_player_by_id")
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
	stmt, err := DB.GetStmt("get_player_by_uuid")
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
	stmt, err := DB.GetStmt("get_player_by_alias")
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

func AddPlayer(DB *db.DB, mcUUID model.UUID, isAllowed bool, isAdmin bool, lastIPV4 *string, aliases []string, invoiceServiceUserId *string) error {
	stmt, err := DB.GetStmt("add_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(mcUUID.Bytes(), isAllowed, isAdmin, lastIPV4, string(aliasesString), invoiceServiceUserId)
	if err != nil {
		return err
	}

	return nil
}

func MakePlayerAliasUnique(DB *db.DB, playerId int64, alias string) error {
	stmt, err := DB.GetStmt("make_player_alias_unique")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(alias, playerId, alias)
	if err != nil {
		return err
	}

	return nil
}

func AddPlayerAlias(DB *db.DB, playerId int64, alias string) error {
	err := MakePlayerAliasUnique(DB, playerId, alias)
	if err != nil {
		return err
	}

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

	player.Aliases = newAliases
	return updatePlayer(DB, player)
}

func UpdatePlayersIsAllowed(DB *db.DB, playerIds []int64, matchIds bool, isAllowed bool) error {
	queryName := "update_players_is_allowed_match"
	if !matchIds {
		queryName = "update_players_is_allowed_not_match"
	}
	stmt, err := DB.GetStmt(queryName)
	if err != nil {
		return err
	}

	playerIdsJson, err := json.Marshal(playerIds)
	if err != nil {
		return fmt.Errorf("error marshaling player IDs: %w", err)
	}

	_, err = stmt.Exec(isAllowed, playerIdsJson)
	if err != nil {
		return err
	}

	return nil
}

func UpdatePlayersIsAdmin(DB *db.DB, playerIds []int64, matchIds bool, isAdmin bool) error {
	queryName := "update_players_is_admin_match"
	if !matchIds {
		queryName = "update_players_is_admin_not_match"
	}
	stmt, err := DB.GetStmt(queryName)
	if err != nil {
		return err
	}

	playerIdsJson, err := json.Marshal(playerIds)
	if err != nil {
		return fmt.Errorf("error marshaling player IDs: %w", err)
	}

	_, err = stmt.Exec(isAdmin, playerIdsJson)
	if err != nil {
		return err
	}

	return nil
}

func UpdatePlayerInvoiceServiceUserId(DB *db.DB, playerId int64, InvoiceServiceUserId *string) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	player.InvoiceServiceUserId = InvoiceServiceUserId
	return updatePlayer(DB, player)
}

func UpdatePlayerIsAllowed(DB *db.DB, playerId int64, isAllowed bool) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	player.IsAllowed = isAllowed
	return updatePlayer(DB, player)
}

func UpdatePlayerIsAdmin(DB *db.DB, playerId int64, isAdmin bool) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	player.IsAdmin = isAdmin
	return updatePlayer(DB, player)
}

func UpdatePlayerLastIPV4(DB *db.DB, playerId int64, lastIPV4 *string) error {
	player, err := GetPlayerById(DB, playerId)
	if err != nil {
		return err
	}

	player.LastIPV4 = lastIPV4
	return updatePlayer(DB, player)
}

func updatePlayer(DB *db.DB, player *model.Player) error {
	stmt, err := DB.GetStmt("update_player")
	if err != nil {
		return err
	}

	aliasesString, err := json.Marshal(player.Aliases)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(player.McUUID.Bytes(), player.IsAllowed, player.IsAdmin, player.LastIPV4, string(aliasesString), player.InvoiceServiceUserId, player.Id)
	if err != nil {
		return err
	}

	return nil
}
