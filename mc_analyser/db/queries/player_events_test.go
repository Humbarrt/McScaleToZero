package queries_test

import (
	"math"
	"math/rand"
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"mc_analyser/model"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayerEvent(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	player_uuid, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoice_service_user_id := "test_invoice_service_user_id"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid, true, true, nil, []string{"Test"}, &invoice_service_user_id)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), player_uuid.Bytes())

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
	player2_uuid, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoice_service_user_id_2 := "test_invoice_service_user_id_2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player2_uuid, true, false, nil, []string{"Test2"}, &invoice_service_user_id_2)
	assert.NoError(t, err)
	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), player2_uuid.Bytes())

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

func TestPlaytimeCalculation(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	player_uuid, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoice_service_user_id := "test_invoice_service_user_id"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid, true, true, nil, []string{"Test"}, &invoice_service_user_id)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), player_uuid.Bytes())

	//Generate some events
	startDate := time.Now().Add(-time.Hour * 24 * 365 * 10).UTC()
	startDate = startDate.Truncate(time.Second)
	currentDate := startDate
	controlPlaytime := time.Duration(0)
	playerEvents := make([]*model.PlayerEvent, 0)
	for i := range 100 {
		next_event_duration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
		next_event_duration = next_event_duration.Truncate(time.Second)
		if i%2 == 1 {
			controlPlaytime += next_event_duration
		}
		currentDate = currentDate.Add(next_event_duration)
		playerEvents = append(playerEvents, &model.PlayerEvent{
			PlayerId:  player.Id,
			Timestamp: currentDate,
			IsOnline:  i%2 == 0,
		})
		err = queries.AddPlayerEvent(DB, currentDate, player.Id, i%2 == 0)
		assert.NoError(t, err)
	}

	//Test if the playtime calculation is correct
	playtime, err := queries.CalculatePlaytime(DB, startDate.Add(-time.Hour*20000), currentDate.Add(time.Hour*20000))
	assert.NoError(t, err)
	assert.Equal(t, controlPlaytime, playtime[0].Playtime)

	//Test if the preOnline calculation is correct (is online since before the startDate)
	durationOnline := (playerEvents[1].Timestamp.Sub(playerEvents[0].Timestamp) / 2).Truncate(time.Second)
	startDate = playerEvents[0].Timestamp.Add(durationOnline)
	preControlPlaytime := controlPlaytime - durationOnline

	playtime, err = queries.CalculatePlaytime(DB, startDate, currentDate.Add(time.Hour*20000))
	assert.NoError(t, err)
	assert.Equal(t, preControlPlaytime, playtime[0].Playtime)

	//Test if the postOnline calculation is correct (still online after the endDate)
	durationOnline = (playerEvents[len(playerEvents)-1].Timestamp.Sub(playerEvents[len(playerEvents)-2].Timestamp) / 2).Truncate(time.Second)
	endDate := playerEvents[len(playerEvents)-2].Timestamp.Add(durationOnline)
	postControlPlaytime := controlPlaytime - durationOnline
	playtime, err = queries.CalculatePlaytime(DB, startDate.Add(-time.Hour*20000), endDate)
	assert.NoError(t, err)
	assert.LessOrEqual(t, postControlPlaytime-1*time.Second, playtime[0].Playtime)
	assert.GreaterOrEqual(t, postControlPlaytime+1*time.Second, playtime[0].Playtime)

	//Test if no player was online in the given time range (before any events)
	playtime, err = queries.CalculatePlaytime(DB, startDate.Add(-time.Hour*40000), startDate.Add(-time.Hour*20000))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if no player was online in the given time range (after all events)
	playtime, err = queries.CalculatePlaytime(DB, currentDate.Add(time.Hour*20000), currentDate.Add(time.Hour*40000))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if no player was online in the given time range (between two sessions)
	testStart := playerEvents[1].Timestamp.Add(time.Second)
	testEnd := playerEvents[2].Timestamp.Add(-time.Second)
	playtime, err = queries.CalculatePlaytime(DB, testStart, testEnd)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(playtime))

	//Test if player is online from before the startDate to after the endDate (during a session)
	testStart = playerEvents[0].Timestamp.Add(time.Second)
	testEnd = playerEvents[1].Timestamp.Add(-time.Second)
	playtime, err = queries.CalculatePlaytime(DB, testStart, testEnd)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(playtime))
	assert.Equal(t, player.Id, playtime[0].PlayerId)
	assert.Equal(t, testEnd.Sub(testStart), playtime[0].Playtime)

	//Test multiple players with interleaved events
	player_uuid2, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoice_service_user_id2 := "test_invoice_service_user_id_2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid2, true, true, nil, []string{"Test2"}, &invoice_service_user_id2)
	assert.NoError(t, err)

	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), player_uuid2.Bytes())

	player_uuid3, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4c")
	invoice_service_user_id3 := "test_invoice_service_user_id_3"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *player_uuid3, true, true, nil, []string{"Test3"}, &invoice_service_user_id3)
	assert.NoError(t, err)

	player3, err := queries.GetPlayerByAlias(DB, "Test3")
	assert.NoError(t, err)
	assert.Equal(t, player3.McUUID.Bytes(), player_uuid3.Bytes())

	//Generate some events
	//playerEvents = make([]*model.PlayerEvent, 0)
	controlPlaytimes := make([]time.Duration, 2)
	for p, playerId := range []int{player2.Id, player3.Id} {
		startDate = time.Now().Add(-time.Hour * 24 * 365 * 10).UTC()
		startDate = startDate.Truncate(time.Second)
		currentDate = startDate
		for i := 0; i < 100; i++ {
			next_event_duration := time.Second * time.Duration(math.Min(36000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600)))))
			next_event_duration = next_event_duration.Truncate(time.Second)
			if i%2 == 1 {
				controlPlaytimes[p] += next_event_duration
			}
			currentDate = currentDate.Add(next_event_duration)
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
	playtime, err = queries.CalculatePlaytime(DB, startDate.Add(-time.Hour*20000), currentDate.Add(time.Hour*20000))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(playtime))
	assert.Equal(t, player.Id, playtime[0].PlayerId)
	assert.Equal(t, player2.Id, playtime[1].PlayerId)
	assert.Equal(t, player3.Id, playtime[2].PlayerId)
	assert.Equal(t, controlPlaytime, playtime[0].Playtime)
	assert.Equal(t, controlPlaytimes[0], playtime[1].Playtime)
	assert.Equal(t, controlPlaytimes[1], playtime[2].Playtime)
}
