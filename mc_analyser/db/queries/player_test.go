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
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid, true, []string{"Test", "TheGoat", "TestDummy"}, "test_splitwise_id1")
	assert.NoError(t, err)

	player_uuid2, err := model.NewUUIDFromString("3cd521d2-c736-40f2-aecc-74362310eac8")
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid2, true, []string{"Test1", "Mustermann"}, "test_splitwise_id2")
	assert.NoError(t, err)

	player_uuid3, err := model.NewUUIDFromString("a1b34617-d5a8-4f75-a67e-b67bd03bd9e4")
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid3, false, []string{"Test2", "Musterfrau"}, "test_splitwise_id3")
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
	err = queries.AddPlayerAlias(DB, 3, "AnotherAlias")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "AnotherAlias")
	assert.NoError(t, err)
	assert.Equal(t, player_uuid3.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, "AnotherAlias", player.Aliases[0])

	//Can set primary alias
	err = queries.AddPlayerAlias(DB, 1, "TestDummy")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "TestDummy")
	assert.NoError(t, err)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())
	assert.Equal(t, "TestDummy", player.Aliases[0])

	//Can set isAllowed
	err = queries.SetPlayerIsAllowed(DB, 1, false)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, false, player.IsAllowed)
	assert.Equal(t, player_uuid.Bytes(), player.McUUID.Bytes())

	//Can set isAllowed
	err = queries.SetPlayerIsAllowed(DB, 3, true)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 3)
	assert.NoError(t, err)
	assert.Equal(t, true, player.IsAllowed)
	assert.Equal(t, player_uuid3.Bytes(), player.McUUID.Bytes())

	//Can set splitwise id
	err = queries.SetPlayerSplitwiseId(DB, 2, "different_splitwise_id")
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, "different_splitwise_id", player.SplitwiseId)
	assert.Equal(t, player_uuid2.Bytes(), player.McUUID.Bytes())
}
