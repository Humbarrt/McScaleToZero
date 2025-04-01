package queries_test

import (
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerEvent(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	err = queries.AddPlayer(DB, "test_mc_id", true, []string{"Test"}, "test_splitwise_id")
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McId, "test_mc_id")

	//If there is no row, we should not be able to add an offline event
	err = queries.AddPlayerEvent(DB, time.Now(), player.Id, false)
	assert.NoError(t, err)
	_, err = queries.GetLastPlayerEvent(DB, player.Id)
	assert.Error(t, err)

	//If the last event is online, we should not be able to add another online event
	time_ref := time.Now().UTC()
	err = queries.AddPlayerEvent(DB, time_ref, player.Id, true)
	assert.NoError(t, err)

	event, err := queries.GetLastPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddPlayerEvent(DB, time.Now(), player.Id, true)
	assert.NoError(t, err)

	event, err = queries.GetLastPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event

	//Check if ordering by timestamp works
	err = queries.AddPlayerEvent(DB, time_ref.Add(-time.Second*10000), player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // It should still be the latest event

	//If the last event is offline, we should be able to add another offline event
	time_ref = time.Now().UTC()
	err = queries.AddPlayerEvent(DB, time_ref, player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddPlayerEvent(DB, time.Now(), player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event

	//Test multiple players online
	err = queries.AddPlayer(DB, "test_mc_id_2", true, []string{"Test2"}, "test_splitwise_id_2")
	assert.NoError(t, err)
	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McId, "test_mc_id_2")

	err = queries.AddPlayerEvent(DB, time.Now(), player.Id, true)
	assert.NoError(t, err)
	err = queries.AddPlayerEvent(DB, time.Now(), player2.Id, true)
	assert.NoError(t, err)

	events, err := queries.GetLastPlayerEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, true)
	assert.Equal(t, events[1].IsOnline, true)

	//Test one player offline/online
	err = queries.AddPlayerEvent(DB, time.Now(), player.Id, false)
	assert.NoError(t, err)

	events, err = queries.GetLastPlayerEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, false)
	assert.Equal(t, events[1].IsOnline, true)

	//Test both players offline
	err = queries.AddPlayerEvent(DB, time.Now(), player2.Id, false)
	assert.NoError(t, err)

	events, err = queries.GetLastPlayerEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, false)
	assert.Equal(t, events[1].IsOnline, false)
}
