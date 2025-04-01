package queries_test

import (
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	err = queries.AddPlayer(DB, "test_mc_id1", true, []string{"Test", "TheGoat", "TestDummy"}, "test_splitwise_id1")
	assert.NoError(t, err)

	err = queries.AddPlayer(DB, "test_mc_id2", true, []string{"Test1", "Mustermann"}, "test_splitwise_id2")
	assert.NoError(t, err)

	err = queries.AddPlayer(DB, "test_mc_id3", false, []string{"Test2", "Musterfrau"}, "test_splitwise_id3")
	assert.NoError(t, err)

	players, err := queries.GetPlayers(DB)
	assert.NoError(t, err)
	assert.Len(t, players, 3)
	assert.Equal(t, "test_mc_id1", players[0].McId)
	assert.Equal(t, "test_mc_id2", players[1].McId)
	assert.Equal(t, "test_mc_id3", players[2].McId)

	player, err := queries.GetPlayerByAlias(DB, "TheGoat")
	assert.NoError(t, err)
	assert.Equal(t, "test_mc_id1", player.McId)

	player, err = queries.GetPlayerById(DB, 3)
	assert.NoError(t, err)
	assert.Equal(t, "test_mc_id3", player.McId)

	//Can add alias to player
	err = queries.AddPlayerAlias(DB, 3, "AnotherAlias")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "AnotherAlias")
	assert.NoError(t, err)
	assert.Equal(t, "test_mc_id3", player.McId)
	assert.Equal(t, "AnotherAlias", player.Aliases[0])

	//Can set primary alias
	err = queries.AddPlayerAlias(DB, 1, "TestDummy")
	assert.NoError(t, err)

	player, err = queries.GetPlayerByAlias(DB, "TestDummy")
	assert.NoError(t, err)
	assert.Equal(t, "test_mc_id1", player.McId)
	assert.Equal(t, "TestDummy", player.Aliases[0])

	//Can set isAllowed
	err = queries.SetPlayerIsAllowed(DB, 1, false)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 1)
	assert.NoError(t, err)
	assert.Equal(t, false, player.IsAllowed)
	assert.Equal(t, "test_mc_id1", player.McId)

	//Can set isAllowed
	err = queries.SetPlayerIsAllowed(DB, 3, true)
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 3)
	assert.NoError(t, err)
	assert.Equal(t, true, player.IsAllowed)
	assert.Equal(t, "test_mc_id3", player.McId)

	//Can set splitwise id
	err = queries.SetPlayerSplitwiseId(DB, 2, "different_splitwise_id")
	assert.NoError(t, err)
	player, err = queries.GetPlayerById(DB, 2)
	assert.NoError(t, err)
	assert.Equal(t, "different_splitwise_id", player.SplitwiseId)
	assert.Equal(t, "test_mc_id2", player.McId)
}
