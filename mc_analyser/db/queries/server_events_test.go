package queries_test

import (
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerEvent(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	//If there is no row, we should not be able to add an offline event
	err = queries.AddServerEvent(DB, time.Now(), false)
	assert.NoError(t, err)
	_, err = queries.GetLastServerEvent(DB)
	assert.Error(t, err)

	//If the last event is online, we should not be able to add another online event
	time_ref := time.Now().UTC()
	err = queries.AddServerEvent(DB, time_ref, true)
	assert.NoError(t, err)

	event, err := queries.GetLastServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddServerEvent(DB, time.Now(), true)
	assert.NoError(t, err)

	event, err = queries.GetLastServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event

	//Check if ordering by timestamp works
	err = queries.AddServerEvent(DB, time_ref.Add(-time.Second*10000), false)
	assert.NoError(t, err)

	event, err = queries.GetLastServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // It should still be the latest event

	//If the last event is offline, we should be able to add another offline event
	time_ref = time.Now().UTC()
	err = queries.AddServerEvent(DB, time_ref, false)
	assert.NoError(t, err)

	event, err = queries.GetLastServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddServerEvent(DB, time.Now(), false)
	assert.NoError(t, err)

	event, err = queries.GetLastServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event
}
