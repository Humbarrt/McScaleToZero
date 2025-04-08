package queries_test

import (
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"mc_analyser/model"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	player_uuid, err := model.NewUUIDFromString("55724e6b-00fd-4e1a-a9e9-224207a57792")
	splitwise_id := "test_splitwise_id1"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid, true, true, nil, []string{"Test", "TheGoat", "TestDummy"}, &splitwise_id)
	assert.NoError(t, err)

	player_uuid2, err := model.NewUUIDFromString("3cd521d2-c736-40f2-aecc-74362310eac8")
	splitwise_id2 := "test_splitwise_id2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid2, true, false, nil, []string{"Test1", "Mustermann"}, &splitwise_id2)
	assert.NoError(t, err)

	player_uuid3, err := model.NewUUIDFromString("a1b34617-d5a8-4f75-a67e-b67bd03bd9e4")
	splitwise_id3 := "test_splitwise_id3"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid3, false, false, nil, []string{"Test2", "Musterfrau"}, &splitwise_id3)
	assert.NoError(t, err)

	players, err := queries.GetPlayers(DB)
	assert.NoError(t, err)
	assert.Len(t, players, 3)
	assert.Equal(t, player_uuid.Bytes(), players[0].McUUID.Bytes())
	assert.Equal(t, player_uuid2.Bytes(), players[1].McUUID.Bytes())
	assert.Equal(t, player_uuid3.Bytes(), players[2].McUUID.Bytes())

	player, err := queries.GetPlayerByAlias(DB, "TheGoat")
	assert.NoError(t, err)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())

	player, err = queries.GetPlayerById(DB, 3)
	assert.NoError(t, err)
	assert.Equal(t, player_uuid3.Bytes(), player.McUUID.Bytes())

	//Can add alias to player
	err = queries.AddPlayerAlias(DB, 3, "Mustermann")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "Mustermann")
	assert.NoError(t, err)
	assert.Equal(t, player_uuid3.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, "Mustermann", player.Aliases[0])

	//Was the alias removed from player 2?
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, player_uuid2.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, 1, len(player.Aliases))
	assert.Equal(t, "Test1", player.Aliases[0])

	//Can i remove all aliases from player 2
	err = queries.MakePlayerAliasUnique(DB, 3, "Test1")
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, player_uuid2.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, 0, len(player.Aliases))
	err = queries.AddPlayerAlias(DB, 2, "Test1")
	assert.NoError(t, err)

	//Can set primary alias
	err = queries.AddPlayerAlias(DB, 1, "TestDummy")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "TestDummy")
	assert.NoError(t, err)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, "TestDummy", player.Aliases[0])

	//Can set isAllowed
	err = queries.UpdatePlayerIsAllowed(DB, 1, false)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, false, player.IsAllowed)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())

	//Can set isAllowed
	err = queries.UpdatePlayerIsAllowed(DB, 3, true)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 3)
	assert.NoError(t, err)
	assert.Equal(t, true, player.IsAllowed)
	assert.Equal(t, player_uuid3.Bytes(), player.McUUID.Bytes())

	//Can set splitwise id
	new_splitwise_id := "different_splitwise_id"
	err = queries.UpdatePlayerSplitwiseId(DB, 2, &new_splitwise_id)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, &new_splitwise_id, player.SplitwiseId)
	assert.Equal(t, player_uuid2.Bytes(), player.McUUID.Bytes())

	//Can set isAdmin
	err = queries.UpdatePlayerIsAdmin(DB, 2, true)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, true, player.IsAdmin)
	assert.Equal(t, player_uuid2.Bytes(), player.McUUID.Bytes())

	err = queries.UpdatePlayerIsAdmin(DB, 1, false)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, false, player.IsAdmin)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())

	//Can set lastIPV4
	newIPV4 := "255.255.255.255"
	err = queries.UpdatePlayerLastIPV4(DB, 1, &newIPV4)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, &newIPV4, player.LastIPV4)

	//Can set lastIPV4 to nil
	err = queries.UpdatePlayerLastIPV4(DB, 1, nil)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Nil(t, player.LastIPV4)

	//Test lastIPV4 length check
	newIPV4 = "255.255.255.255."
	err = queries.UpdatePlayerLastIPV4(DB, 1, &newIPV4)
	assert.Error(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Nil(t, player.LastIPV4)

	//Can get player by UUID
	player, err = queries.GetPlayerByUUID(DB, *player_uuid)
	assert.NoError(t, err)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())

	//Can not set same UUID
	err = queries.AddPlayer(DB, *player_uuid, true, true, nil, []string{"Test", "TheGoat", "TestDummy"}, &splitwise_id)
	assert.Error(t, err)

	//Bulk update isAllowed for all players that match the ids
	ids := []int{1, 2}
	err = queries.UpdatePlayersIsAllowed(DB, ids, true, false)
	assert.NoError(t, err)
	expectPlayersIsAllowed(t, DB, map[int]bool{
		1: false,
		2: false,
		3: true,
	})
	err = queries.UpdatePlayersIsAllowed(DB, ids, true, true)
	assert.NoError(t, err)
	expectPlayersIsAllowed(t, DB, map[int]bool{
		1: true,
		2: true,
		3: true,
	})

	//Bulk update isAllowed for all players that do not match the ids
	err = queries.UpdatePlayersIsAllowed(DB, ids, false, false)
	assert.NoError(t, err)
	expectPlayersIsAllowed(t, DB, map[int]bool{
		1: true,
		2: true,
		3: false,
	})
	ids = []int{3}
	err = queries.UpdatePlayersIsAllowed(DB, ids, false, false)
	assert.NoError(t, err)
	expectPlayersIsAllowed(t, DB, map[int]bool{
		1: false,
		2: false,
		3: false,
	})

	//Bulk update isAdmin for all players that match the ids
	ids = []int{1, 2}
	err = queries.UpdatePlayersIsAdmin(DB, ids, true, false)
	assert.NoError(t, err)
	expectPlayersIsAdmin(t, DB, map[int]bool{
		1: false,
		2: false,
		3: false,
	})
	err = queries.UpdatePlayersIsAdmin(DB, ids, true, true)
	assert.NoError(t, err)
	expectPlayersIsAdmin(t, DB, map[int]bool{
		1: true,
		2: true,
		3: false,
	})

	//Bulk update isAdmin for all players that do not match the ids
	err = queries.UpdatePlayersIsAdmin(DB, ids, false, true)
	assert.NoError(t, err)
	expectPlayersIsAdmin(t, DB, map[int]bool{
		1: true,
		2: true,
		3: true,
	})
	ids = []int{3}
	err = queries.UpdatePlayersIsAdmin(DB, ids, false, false)
	assert.NoError(t, err)
	expectPlayersIsAdmin(t, DB, map[int]bool{
		1: false,
		2: false,
		3: true,
	})
}

func expectPlayersIsAllowed(t *testing.T, DB *db.DB, pattern map[int]bool) {
	for id, isAllowed := range pattern {
		player, err := queries.GetPlayerById(DB, id)
		assert.NoError(t, err)
		assert.Equal(t, isAllowed, player.IsAllowed)
	}
}

func expectPlayersIsAdmin(t *testing.T, DB *db.DB, pattern map[int]bool) {
	for id, isAllowed := range pattern {
		player, err := queries.GetPlayerById(DB, id)
		assert.NoError(t, err)
		assert.Equal(t, isAllowed, player.IsAdmin)
	}
}
