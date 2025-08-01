package queries_test

import (
	"math"
	"math/rand"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/model"
	"path"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerEvent(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	playerUUID, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoiceServiceUserId := "testInvoiceServiceUserId"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID, true, true, nil, []string{"Test"}, &invoiceServiceUserId)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), playerUUID.Bytes())

	//If there is no row, we should not be able to add an offline event
	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player.Id, false)
	assert.NoError(t, err)
	_, err = queries.GetLastestPlayerEvent(DB, player.Id)
	assert.Error(t, err)

	//If the last event is online, we should not be able to add another online event
	timeRef := time.Now().UTC()
	err = queries.AddPlayerEvent(DB, timeRef, player.Id, true)
	assert.NoError(t, err)

	event, err := queries.GetLastestPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, timeRef)

	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player.Id, true)
	assert.NoError(t, err)

	event, err = queries.GetLastestPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, timeRef) // There should be no new event

	//Check if ordering by timestamp works
	err = queries.AddPlayerEvent(DB, timeRef.Add(-time.Second*10000), player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastestPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, true)
	assert.Equal(t, event.Timestamp, timeRef) // It should still be the latest event

	//If the last event is offline, we should be able to add another offline event
	timeRef = time.Now().UTC()
	err = queries.AddPlayerEvent(DB, timeRef, player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastestPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, timeRef)

	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player.Id, false)
	assert.NoError(t, err)

	event, err = queries.GetLastestPlayerEvent(DB, player.Id)
	assert.NoError(t, err)
	assert.Equal(t, event.IsOnline, false)
	assert.Equal(t, event.Timestamp, timeRef) // There should be no new event

	//Test multiple players online
	player2UUID, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoiceServiceUserId2 := "testInvoiceServiceUserId2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player2UUID, true, false, nil, []string{"Test2"}, &invoiceServiceUserId2)
	assert.NoError(t, err)
	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), player2UUID.Bytes())

	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player.Id, true)
	assert.NoError(t, err)
	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player2.Id, true)
	assert.NoError(t, err)

	events, err := queries.GetLastestPlayersEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, true)
	assert.Equal(t, events[1].IsOnline, true)

	//Test one player offline/online
	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player.Id, false)
	assert.NoError(t, err)

	events, err = queries.GetLastestPlayersEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, false)
	assert.Equal(t, events[1].IsOnline, true)

	//Test both players offline
	err = queries.AddPlayerEvent(DB, time.Now().UTC(), player2.Id, false)
	assert.NoError(t, err)

	events, err = queries.GetLastestPlayersEvents(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(events), 2)
	assert.Equal(t, events[0].PlayerId, player.Id)
	assert.Equal(t, events[1].PlayerId, player2.Id)
	assert.Equal(t, events[0].IsOnline, false)
	assert.Equal(t, events[1].IsOnline, false)
}

func TestPlaytimeCalculation(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	playerUUID, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoiceServiceUserId := "testInvoiceServiceUserId"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID, true, true, nil, []string{"Test"}, &invoiceServiceUserId)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), playerUUID.Bytes())

	//Generate some events
	startDate := time.Now().UTC().Add(-time.Hour * 24 * 365 * 10).UTC()
	startDate = startDate.Truncate(time.Second)
	currentDate := startDate
	controlPlaytime := time.Duration(0)
	playerEvents := make([]*model.PlayerEvent, 0)
	for i := range 100 {
		nextEventDuration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
		nextEventDuration = nextEventDuration.Truncate(time.Second)
		if i%2 == 1 {
			controlPlaytime += nextEventDuration
		}
		currentDate = currentDate.Add(nextEventDuration)
		playerEvents = append(playerEvents, &model.PlayerEvent{
			PlayerId:  player.Id,
			Timestamp: currentDate,
			IsOnline:  i%2 == 0,
		})
		err = queries.AddPlayerEvent(DB, currentDate, player.Id, i%2 == 0)
		assert.NoError(t, err)
	}

	//Test if the playtime calculation is correct
	playtime, err := queries.CalculatePlaytime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 20000), End: currentDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, controlPlaytime, playtime[player.Id])

	//Test if the preOnline calculation is correct (is online since before the startDate)
	durationOnline := (playerEvents[1].Timestamp.Sub(playerEvents[0].Timestamp) / 2).Truncate(time.Second)
	startDate = playerEvents[0].Timestamp.Add(durationOnline)
	preControlPlaytime := controlPlaytime - durationOnline

	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: startDate, End: currentDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, preControlPlaytime, playtime[player.Id])

	//Test if the postOnline calculation is correct (still online after the endDate)
	durationOnline = (playerEvents[len(playerEvents)-1].Timestamp.Sub(playerEvents[len(playerEvents)-2].Timestamp) / 2).Truncate(time.Second)
	endDate := playerEvents[len(playerEvents)-2].Timestamp.Add(durationOnline)
	postControlPlaytime := controlPlaytime - durationOnline
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 20000), End: endDate})
	assert.NoError(t, err)
	assert.LessOrEqual(t, postControlPlaytime-1*time.Second, playtime[player.Id])
	assert.GreaterOrEqual(t, postControlPlaytime+1*time.Second, playtime[player.Id])

	//Test if no player was online in the given time range (before any events)
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 40000), End: startDate.Add(-time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if no player was online in the given time range (after all events)
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: currentDate.Add(time.Hour * 20000), End: currentDate.Add(time.Hour * 40000)})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if no player was online in the given time range (between two sessions)
	testStart := playerEvents[1].Timestamp.Add(time.Second)
	testEnd := playerEvents[2].Timestamp.Add(-time.Second)
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: testStart, End: testEnd})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if player is online from before the startDate to after the endDate (during a session)
	testStart = playerEvents[0].Timestamp.Add(time.Second)
	testEnd = playerEvents[1].Timestamp.Add(-time.Second)
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: testStart, End: testEnd})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(playtime))
	assert.Equal(t, testEnd.Sub(testStart), playtime[player.Id])

	//Test multiple players with interleaved events
	playerUUID2, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoiceServiceUserId2 := "testInvoiceServiceUserId2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID2, true, true, nil, []string{"Test2"}, &invoiceServiceUserId2)
	assert.NoError(t, err)

	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), playerUUID2.Bytes())

	playerUUID3, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4c")
	invoiceServiceUserId3 := "testInvoiceServiceUserId3"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID3, true, true, nil, []string{"Test3"}, &invoiceServiceUserId3)
	assert.NoError(t, err)

	player3, err := queries.GetPlayerByAlias(DB, "Test3")
	assert.NoError(t, err)
	assert.Equal(t, player3.McUUID.Bytes(), playerUUID3.Bytes())

	//Generate some events
	//playerEvents = make([]*model.PlayerEvent, 0)
	controlPlaytimes := make([]time.Duration, 2)
	for p, playerId := range []int64{player2.Id, player3.Id} {
		startDate = time.Now().UTC().Add(-time.Hour * 24 * 365 * 10).UTC()
		startDate = startDate.Truncate(time.Second)
		currentDate = startDate
		for i := 0; i < 100; i++ {
			nextEventDuration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
			nextEventDuration = nextEventDuration.Truncate(time.Second)
			if i%2 == 1 {
				controlPlaytimes[p] += nextEventDuration
			}
			currentDate = currentDate.Add(nextEventDuration)
			/* playerEvents = append(playerEvents, &model.PlayerEvent{
				PlayerId:  playerId,
				Timestamp: currentDate,
				IsOnline:  i%2 == 0,
			}) */
			err = queries.AddPlayerEvent(DB, currentDate, playerId, i%2 == 0)
			assert.NoError(t, err)
		}
	}

	//Test if the playtime calculation is correct (if adding more players changes the result)
	playtime, err = queries.CalculatePlaytime(DB, model.Timeframe{Start: startDate.Add(-time.Hour * 20000), End: currentDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(playtime))
	assert.Equal(t, controlPlaytime, playtime[player.Id])
	assert.Equal(t, controlPlaytimes[0], playtime[player2.Id])
	assert.Equal(t, controlPlaytimes[1], playtime[player3.Id])
}

func TestCPUTimeCalculation(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	playerUUID, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoiceServiceUserId := "testInvoiceServiceUserId"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID, true, true, nil, []string{"Test"}, &invoiceServiceUserId)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), playerUUID.Bytes())

	playerUUID2, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoiceServiceUserId2 := "testInvoiceServiceUserId2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID2, true, true, nil, []string{"Test2"}, &invoiceServiceUserId2)
	assert.NoError(t, err)

	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), playerUUID2.Bytes())

	playerUUID3, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4c")
	invoiceServiceUserId3 := "testInvoiceServiceUserId3"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID3, true, true, nil, []string{"Test3"}, &invoiceServiceUserId3)
	assert.NoError(t, err)

	player3, err := queries.GetPlayerByAlias(DB, "Test3")
	assert.NoError(t, err)
	assert.Equal(t, player3.McUUID.Bytes(), playerUUID3.Bytes())

	//Generate some events
	//Test multiple players with interleaved events

	events := make([]*model.PlayerEvent, 0)
	playerIds := []int64{player.Id, player2.Id, player3.Id}
	initDate := time.Now().UTC().Add(-time.Hour * 24 * 365 * 10).UTC()
	lastDate := initDate

	for _, playerId := range playerIds {
		startDate := initDate
		startDate = startDate.Truncate(time.Second)
		currentDate := startDate
		for i := range 100 {
			nextEventDuration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
			nextEventDuration = nextEventDuration.Truncate(time.Second)
			currentDate = currentDate.Add(nextEventDuration)
			events = append(events, &model.PlayerEvent{
				PlayerId:  playerId,
				Timestamp: currentDate,
				IsOnline:  i%2 == 0,
			})
			err = queries.AddPlayerEvent(DB, currentDate, playerId, i%2 == 0)
			assert.NoError(t, err)
		}
		if currentDate.After(lastDate) {
			lastDate = currentDate
		}
	}

	//sort the events by timestamp, earliest first
	controlCPUtimes := make([]time.Duration, 3)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	//Calculate the CPU time for each player
	var playersOnline int64
	playerStates := map[int64]bool{player.Id: false, player2.Id: false, player3.Id: false}
	for e, event := range events {
		for p := range playerIds {
			if playerStates[playerIds[p]] {
				timeSinceLastEvent := events[e].Timestamp.Sub(events[e-1].Timestamp)
				cpuTimeSinceLastEvent := timeSinceLastEvent.Nanoseconds() / playersOnline
				controlCPUtimes[p] += time.Duration(cpuTimeSinceLastEvent)
			}
		}

		playerStates[event.PlayerId] = event.IsOnline
		if event.IsOnline {
			playersOnline++
		} else {
			playersOnline--
		}
	}

	totalCPUTime, err := queries.CalculateCPUTime(DB, model.Timeframe{
		Start: initDate.Add(-time.Hour * 20000),
		End:   events[len(events)-1].Timestamp.Add(time.Hour * 20000),
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(totalCPUTime))

	for playerId, duration := range totalCPUTime {
		for p := range playerIds {
			if playerId != playerIds[p] {
				continue
			}
			assert.Equal(t, controlCPUtimes[p], duration)
			break
		}
	}

	playtime, err := queries.CalculatePlaytime(DB, model.Timeframe{Start: initDate.Add(-time.Hour * 20000), End: initDate.Add(time.Hour * 20000)})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(playtime))

	//Test if the preOnline calculation is correct (is online since before the startDate)
	durationOnline := (events[1].Timestamp.Sub(events[0].Timestamp) / 2).Truncate(time.Second)
	startDate := events[0].Timestamp.Add(durationOnline)
	var preControlCPUtime time.Duration
	for playerId, duration := range totalCPUTime {
		if playerId != events[0].PlayerId {
			continue
		}
		preControlCPUtime = duration
		break
	}
	preControlCPUtime -= durationOnline

	cpuTime, err := queries.CalculateCPUTime(DB, model.Timeframe{
		Start: startDate, End: lastDate.Add(time.Hour * 20000),
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))

	wasFound := false
	for playerId, duration := range cpuTime {
		if playerId != events[0].PlayerId {
			continue
		}
		assert.Equal(t, preControlCPUtime, duration)
		wasFound = true
	}
	assert.True(t, wasFound, "Player %d not found in pre calculation", events[0].PlayerId)

	//Test if the postOnline calculation is correct (still online after the endDate)
	durationOnline = (events[len(events)-1].Timestamp.Sub(events[len(events)-2].Timestamp) / 2).Truncate(time.Second)
	endDate := events[len(events)-2].Timestamp.Add(durationOnline)
	var postControlCPUtime time.Duration
	for playerId, duration := range totalCPUTime {
		if playerId != events[len(events)-1].PlayerId {
			continue
		}
		postControlCPUtime = duration
		break
	}
	postControlCPUtime -= durationOnline
	cpuTime, err = queries.CalculateCPUTime(DB, model.Timeframe{
		Start: startDate.Add(-time.Hour * 20000),
		End:   endDate,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))

	wasFound = false
	for playerId, duration := range cpuTime {
		if playerId != events[len(events)-1].PlayerId {
			continue
		}
		assert.GreaterOrEqual(t, postControlCPUtime+time.Duration(1)*time.Second, duration)
		assert.LessOrEqual(t, postControlCPUtime-time.Duration(1)*time.Second, duration)
		wasFound = true
	}
	assert.True(t, wasFound, "Player %d not found in post calculation", events[len(events)-1].PlayerId)

	//Test if no player was online in the given time range (before any events)
	cpuTime, err = queries.CalculateCPUTime(DB, model.Timeframe{
		Start: initDate.Add(-time.Hour * 40000), End: initDate.Add(-time.Hour * 20000),
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))
	for _, duration := range cpuTime {
		assert.Equal(t, time.Duration(0), duration)
	}

	//Test if no player was online in the given time range (after all events)
	cpuTime, err = queries.CalculateCPUTime(DB, model.Timeframe{
		Start: lastDate.Add(time.Hour * 20000), End: lastDate.Add(time.Hour * 40000),
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))
	for _, duration := range cpuTime {
		assert.Equal(t, time.Duration(0), duration)
	}

	//Test if no player was online in the given time range (between two sessions)
	eventId := -1
	playersOnline = 0
	for e := range events {
		if events[e].IsOnline {
			playersOnline++
		} else {
			playersOnline--
		}

		if e != 0 && playersOnline == 0 {
			eventId = e
			break
		}

		if e == len(events)-1 {
			t.Fatal("No event found with players online")
		}
	}

	startDate = events[eventId].Timestamp.Add(time.Second)
	endDate = events[eventId+1].Timestamp.Add(-time.Second)
	cpuTime, err = queries.CalculateCPUTime(DB, model.Timeframe{
		Start: startDate, End: endDate,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))
	for playerId := range cpuTime {
		assert.Equal(t, time.Duration(0), cpuTime[playerId])
	}

	//Test if player is online from before the startDate to after the endDate (during a session)
	startDate = events[0].Timestamp.Add(time.Second)
	endDate = events[1].Timestamp.Add(-time.Second)
	cpuTime, err = queries.CalculateCPUTime(DB, model.Timeframe{
		Start: startDate, End: endDate,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cpuTime))
	for playerId, duration := range cpuTime {
		if playerId != events[0].PlayerId {
			continue
		}
		assert.Equal(t, endDate.Sub(startDate), duration)
		break
	}
	assert.True(t, wasFound, "Player %d not found in online calculation", events[0].PlayerId)
}
