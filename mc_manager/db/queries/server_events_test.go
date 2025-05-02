package queries_test

import (
	"math"
	"math/rand"
	"mc_manager/db"
	"mc_manager/model"
	"mc_managerdb/queries"
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
	err = queries.AddServerEvent(DB, time.Now().UTC(), false)
	assert.NoError(t, err)
	_, err = queries.GetLatestServerEvent(DB)
	assert.Error(t, err)

	//If the last event is online, we should not be able to add another online event
	time_ref := time.Now().UTC()
	err = queries.AddServerEvent(DB, time_ref, true)
	assert.NoError(t, err)

	event, err := queries.GetLatestServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddServerEvent(DB, time.Now().UTC(), true)
	assert.NoError(t, err)

	event, err = queries.GetLatestServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event

	//Check if ordering by timestamp works
	err = queries.AddServerEvent(DB, time_ref.Add(-time.Second*10000), false)
	assert.NoError(t, err)

	event, err = queries.GetLatestServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, time_ref) // It should still be the latest event

	//If the last event is offline, we should be able to add another offline event
	time_ref = time.Now().UTC()
	err = queries.AddServerEvent(DB, time_ref, false)
	assert.NoError(t, err)

	event, err = queries.GetLatestServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref)

	err = queries.AddServerEvent(DB, time.Now().UTC(), false)
	assert.NoError(t, err)

	event, err = queries.GetLatestServerEvent(DB)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, time_ref) // There should be no new event
}

func TestServerUptimeCalculation(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	//Generate some events
	startDate := time.Now().UTC().Add(-time.Hour * 24 * 365 * 10).UTC()
	startDate = startDate.Truncate(time.Second)
	currentDate := startDate
	controlUptime := time.Duration(0)
	serverEvents := make([]*model.ServerEvent, 0)
	for i := range 100 {
		next_event_duration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
		next_event_duration = next_event_duration.Truncate(time.Second)
		if i%2 == 1 {
			controlUptime += next_event_duration
		}
		currentDate = currentDate.Add(next_event_duration)
		serverEvents = append(serverEvents, &model.ServerEvent{
			Timestamp: currentDate,
			IsOnline:  i%2 == 0,
		})
		err = queries.AddServerEvent(DB, currentDate, i%2 == 0)
		assert.NoError(t, err)
	}

	//Test if the uptime calculation is correct
	uptime, err := queries.CalculateServerUptime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 20000), End: currentDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, controlUptime, uptime)

	//Test if the preOnline calculation is correct (is online since before the startDate)
	durationOnline := (serverEvents[1].Timestamp.Sub(serverEvents[0].Timestamp) / 2).Truncate(time.Second)
	startDate = serverEvents[0].Timestamp.Add(durationOnline)
	precontrolUptime := controlUptime - durationOnline

	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: startDate, End: currentDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, precontrolUptime, uptime)

	//Test if the postOnline calculation is correct (still online after the endDate)
	durationOnline = (serverEvents[len(serverEvents)-1].Timestamp.Sub(serverEvents[len(serverEvents)-2].Timestamp) / 2).Truncate(time.Second)
	endDate := serverEvents[len(serverEvents)-2].Timestamp.Add(durationOnline)
	postcontrolUptime := controlUptime - durationOnline
	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 20000), End: endDate})
	assert.NoError(t, err)
	assert.LessOrEqual(t, postcontrolUptime-1*time.Second, uptime)
	assert.GreaterOrEqual(t, postcontrolUptime+1*time.Second, uptime)

	//Test if never online in the given time range (before any events)
	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 40000), End: startDate.Add(-time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), uptime)

	//Test if never online in the given time range (after all events)
	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: currentDate.Add(time.Hour * 20000), End: currentDate.Add(time.Hour * 40000)})
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), uptime)

	//Test if never online in the given time range (between two sessions)
	testStart := serverEvents[1].Timestamp.Add(time.Second)
	testEnd := serverEvents[2].Timestamp.Add(-time.Second)
	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: testStart, End: testEnd})
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), uptime)

	//Test if is online from before the startDate to after the endDate (during a session)
	testStart = serverEvents[0].Timestamp.Add(time.Second)
	testEnd = serverEvents[1].Timestamp.Add(-time.Second)
	uptime, err = queries.CalculateServerUptime(DB, model.Timeframe{Start: testStart, End: testEnd})
	assert.NoError(t, err)
	assert.Equal(t, testEnd.Sub(testStart), uptime)
}
