package server_watcher

import (
	"bufio"
	"encoding/json"
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"mc_analyser/model"
	"path"
	"strings"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestProcessMcPlayers(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	uuids := []model.UUID{}
	for i := 1; i <= 10; i++ {
		b := make([]byte, 16)
		b[0] = byte(i)

		uuid, err := model.NewUUIDFromBytes(b)
		assert.NoError(t, err)
		uuids = append(uuids, *uuid)
	}

	whitelist := []*mcPlayer{}
	for _, i := range []int{0, 2, 3, 5, 6, 8} {
		whitelist = append(whitelist, &mcPlayer{
			UUID: uuids[i],
			Name: "Test" + fmt.Sprint(i),
		})
	}
	whitelist = append(whitelist, &mcPlayer{
		UUID: uuids[2],
		Name: "ThisShouldbeSecond",
	})

	admins := []*mcPlayer{}
	for _, i := range []int{1, 4, 5, 7, 9} {
		admins = append(admins, &mcPlayer{
			UUID: uuids[i],
			Name: "Test" + fmt.Sprint(i),
		})
	}
	admins = append(admins, &mcPlayer{
		UUID: uuids[2],
		Name: "ThisShouldBeFirst",
	})

	err = processMcPlayers(DB, whitelist, admins)
	assert.NoError(t, err)

	players, err := queries.GetPlayers(DB)
	assert.NoError(t, err)
	assert.Len(t, players, 10)
	expectedPlayers := []*model.Player{
		{Id: 1, McUUID: uuids[0], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test0"}},                                           //was in whitelist, dropped
		{Id: 7, McUUID: uuids[1], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test1"}},                                            //was in admin, dropped
		{Id: 2, McUUID: uuids[2], IsAllowed: true, IsAdmin: true, Aliases: []string{"ThisShouldBeFirst", "ThisShouldbeSecond", "Test2"}}, //was in both, dropped
		{Id: 3, McUUID: uuids[3], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test3"}},                                           //was in whitelist, kept
		{Id: 8, McUUID: uuids[4], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test4"}},                                            //was in admin, kept
		{Id: 4, McUUID: uuids[5], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test5"}},                                            //was in both, kept
		{Id: 5, McUUID: uuids[6], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test6"}},                                           //was in whitelist, now both
		{Id: 9, McUUID: uuids[7], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test7"}},                                            //was in admin, now both
		{Id: 6, McUUID: uuids[8], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test8"}},                                           //was in whitelist, now only admin
		{Id: 10, McUUID: uuids[9], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test9"}},                                           //was in admin, now only whitelist
	}
	for _, ePlayer := range expectedPlayers {
		player := players[ePlayer.Id-1]
		assert.Equal(t, ePlayer.Id, player.Id)
		assert.Equal(t, ePlayer.McUUID, player.McUUID)
		assert.Equal(t, ePlayer.IsAllowed, player.IsAllowed)
		assert.Equal(t, ePlayer.IsAdmin, player.IsAdmin)
		assert.Equal(t, ePlayer.Aliases, player.Aliases)
	}

	// Test changing whitelist and admins
	whitelist = []*mcPlayer{}
	for _, i := range []int{3, 5, 6, 7, 9} {
		whitelist = append(whitelist, &mcPlayer{
			UUID: uuids[i],
			Name: "Test" + fmt.Sprint(i),
		})
	}

	admins = []*mcPlayer{}
	for _, i := range []int{4, 5, 6, 7, 8} {
		admins = append(admins, &mcPlayer{
			UUID: uuids[i],
			Name: "Test" + fmt.Sprint(i),
		})
	}

	err = processMcPlayers(DB, whitelist, admins)
	assert.NoError(t, err)

	players, err = queries.GetPlayers(DB)
	assert.NoError(t, err)
	assert.Len(t, players, 10)

	expectedPlayers = []*model.Player{
		{Id: 1, McUUID: uuids[0], IsAllowed: false, IsAdmin: false, Aliases: []string{"Test0"}},                                            //was in whitelist, dropped
		{Id: 7, McUUID: uuids[1], IsAllowed: false, IsAdmin: false, Aliases: []string{"Test1"}},                                            //was in admin, dropped
		{Id: 2, McUUID: uuids[2], IsAllowed: false, IsAdmin: false, Aliases: []string{"ThisShouldBeFirst", "ThisShouldbeSecond", "Test2"}}, //was in both, dropped
		{Id: 3, McUUID: uuids[3], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test3"}},                                             //was in whitelist, kept
		{Id: 8, McUUID: uuids[4], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test4"}},                                              //was in admin, kept
		{Id: 4, McUUID: uuids[5], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test5"}},                                              //was in both, kept
		{Id: 5, McUUID: uuids[6], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test6"}},                                              //was in whitelist, now both
		{Id: 9, McUUID: uuids[7], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test7"}},                                              //was in admin, now both
		{Id: 6, McUUID: uuids[8], IsAllowed: true, IsAdmin: true, Aliases: []string{"Test8"}},                                              //was in whitelist, now only admin
		{Id: 10, McUUID: uuids[9], IsAllowed: true, IsAdmin: false, Aliases: []string{"Test9"}},                                            //was in admin, now only whitelist
	}
	for _, ePlayer := range expectedPlayers {
		player := players[ePlayer.Id-1]
		assert.Equal(t, ePlayer.Id, player.Id)
		assert.Equal(t, ePlayer.McUUID, player.McUUID)
		assert.Equal(t, ePlayer.IsAllowed, player.IsAllowed)
		assert.Equal(t, ePlayer.IsAdmin, player.IsAdmin)
		assert.Equal(t, ePlayer.Aliases, player.Aliases)
	}
}

func TestParsingMcFile(t *testing.T) {
	uuids := []model.UUID{}
	for i := 1; i <= 10; i++ {
		b := make([]byte, 16)
		b[0] = byte(i)

		uuid, err := model.NewUUIDFromBytes(b)
		assert.NoError(t, err)
		uuids = append(uuids, *uuid)
	}
	mcPlayers := []*mcPlayer{}
	for _, i := range []int{0, 2, 3, 5, 6, 8} {
		mcPlayers = append(mcPlayers, &mcPlayer{
			UUID: uuids[i],
			Name: "Test" + fmt.Sprint(i),
		})
	}

	mcPlayersJson, err := json.Marshal(mcPlayers)
	assert.NoError(t, err)

	buf := bufio.NewReader(strings.NewReader(string(mcPlayersJson)))
	parsedMcPlayers, err := parseMcPlayerFile(buf)
	assert.NoError(t, err)
	assert.Len(t, parsedMcPlayers, len(mcPlayers))
	for i, mcPlayer := range parsedMcPlayers {
		assert.Equal(t, mcPlayers[i].UUID, mcPlayer.UUID)
		assert.Equal(t, mcPlayers[i].Name, mcPlayer.Name)
	}

}
