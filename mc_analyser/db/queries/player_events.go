package queries

import (
	"database/sql"
	"fmt"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
	"time"
)

func parsePlayerEvent(row scannable) (*model.PlayerEvent, error) {
	var id int64
	var timestamp time.Time
	var playerId int64
	var isOnline bool
	err := row.Scan(&id, &timestamp, &playerId, &isOnline)
	if err != nil {
		return nil, err
	}

	return &model.PlayerEvent{
		Id:        id,
		PlayerId:  playerId,
		Timestamp: timestamp,
		IsOnline:  isOnline,
	}, nil
}

func AddPlayerEvent(DB *db.DB, timestamp time.Time, player_id int64, isOnline bool) error {
	lastState, err := GetLastestPlayerEvent(DB, player_id)
	if err == sql.ErrNoRows {
		lastState = &model.PlayerEvent{
			Id:        0,
			PlayerId:  player_id,
			Timestamp: time.Time{},
			IsOnline:  false,
		}
	} else if err != nil {
		return err
	}

	if lastState.IsOnline == isOnline || !lastState.Timestamp.IsZero() && timestamp.Before(lastState.Timestamp) {
		return nil
	}

	stmt, err := DB.GetStmt("add_player_event")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(timestamp, player_id, isOnline)
	return err
}

func GetLastestPlayerEvent(DB *db.DB, playerId int64) (*model.PlayerEvent, error) {
	stmt, err := DB.GetStmt("get_latest_player_event")
	if err != nil {
		return &model.PlayerEvent{}, err
	}

	result := stmt.QueryRow(playerId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parsePlayerEvent(result)
}

func GetLastPlayersEvents(DB *db.DB, before time.Time) ([]*model.PlayerEvent, error) {
	stmt, err := DB.GetStmt("get_last_players_events")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(
		sql.Named("before_time", before),
	)
	if err != nil {
		return nil, err
	}

	events := []*model.PlayerEvent{}
	for result.Next() {
		event, err := parsePlayerEvent(result)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func GetLastestPlayersEvents(DB *db.DB) ([]*model.PlayerEvent, error) {
	return GetLastPlayersEvents(DB, time.Now().UTC())
}

func GetPlayersEvents(DB *db.DB, timeframe model.Timeframe) ([]*model.PlayerEvent, error) {
	stmt, err := DB.GetStmt("get_players_events")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(
		sql.Named("start_time", timeframe.Start),
		sql.Named("end_time", timeframe.End),
	)
	if err != nil {
		return nil, err
	}

	events := []*model.PlayerEvent{}
	for result.Next() {
		event, err := parsePlayerEvent(result)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func CalculatePlaytime(DB *db.DB, timeframe model.Timeframe) (map[int64]time.Duration, error) {
	calcStart := time.Now().UTC()
	stmt, err := DB.GetStmt("calculate_playtime")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(
		sql.Named("start_time", timeframe.Start),
		sql.Named("end_time", timeframe.End),
	)
	if err != nil {
		return nil, err
	}

	totalPlaytime := map[int64]time.Duration{}
	for result.Next() {
		var playerId int64
		var playtimeInSec int64
		err := result.Scan(&playerId, &playtimeInSec)
		if err != nil {
			return nil, err
		}

		playtime := time.Duration(playtimeInSec) * time.Second
		totalPlaytime[playerId] = playtime
	}
	logging.Debug("Playtime calculation took", "duration", time.Since(calcStart))
	return totalPlaytime, nil
}

func CalculateCPUTime(DB *db.DB, timeframe model.Timeframe) (map[int64]time.Duration, error) {
	calcStart := time.Now().UTC()
	events, err := GetPlayersEvents(DB, timeframe)
	if err != nil {
		logging.Error("Error getting player events", "error", err)
		return nil, err
	}

	playerIds := []int64{}
	playerStates := make(map[int64]bool)
	cpuTimes := map[int64]time.Duration{}
	players, err := GetPlayers(DB)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		return nil, err
	}

	for _, player := range players {
		playerIds = append(playerIds, player.Id)
		cpuTimes[player.Id] = 0
		playerStates[player.Id] = false
	}

	//get initial player states
	var playersOnline int64
	playerEventsBefore, err := GetLastPlayersEvents(DB, timeframe.Start)
	if err != nil {
		logging.Error("Error getting player events before start", "error", err)
		return nil, err
	}

	//get the last player events before start
	for _, event := range playerEventsBefore {
		if event.Timestamp.After(timeframe.Start) {
			return nil, fmt.Errorf("event timestamp %s is after start %s", event.Timestamp, timeframe.Start)
		}

		playerStates[event.PlayerId] = event.IsOnline
		if event.IsOnline {
			playersOnline++
		}
	}

	//Calculate the CPU time for each player
	for e, event := range events {
		/* log.Println("--------------------------")
		log.Println("players online before: ", playersOnline)
		if e > 0 {
			log.Println("time since last event: ", events[e].Timestamp.Sub(events[e-1].Timestamp).Seconds())
		}
		log.Println("event: player_", events[e].PlayerId, " isOnline: ", event.IsOnline) */

		for _, playerId := range playerIds {
			if !playerStates[playerId] {
				continue
			}

			lastEventTimestamp := timeframe.Start
			if e != 0 {
				lastEventTimestamp = events[e-1].Timestamp
			}

			timeSinceLastEvent := events[e].Timestamp.Sub(lastEventTimestamp)
			cpuTimeSinceLastEvent := timeSinceLastEvent.Nanoseconds() / playersOnline
			cpuTimes[playerId] += time.Duration(cpuTimeSinceLastEvent)

			//log.Println("player_", playerIds[p], " cpuTimeSinceLastEvent: ", cpuTimeSinceLastEvent, " -> total: ", cpuTimes[p].Seconds())
		}

		playerStates[event.PlayerId] = event.IsOnline
		if event.IsOnline {
			playersOnline++
		} else {
			playersOnline--
		}
	}

	for _, playerId := range playerIds {
		if !playerStates[playerId] {
			continue
		}

		/* log.Println("--------------------------")
		log.Println("player_", playerId, "session did not end") */

		lastEventTimestamp := timeframe.Start
		if len(events) != 0 {
			lastEventTimestamp = events[len(events)-1].Timestamp
		}

		timeSinceLastEvent := timeframe.End.Sub(lastEventTimestamp)
		cpuTimeSinceLastEvent := timeSinceLastEvent.Nanoseconds() / playersOnline
		cpuTimes[playerId] += time.Duration(cpuTimeSinceLastEvent)

		//log.Println("end player_", playerId, " cpuTimeSinceLastEvent: ", cpuTimeSinceLastEvent, " -> total: ", cpuTimes[playerId].Seconds())
	}

	logging.Debug("CPU time calculation took", "duration", time.Since(calcStart))
	return cpuTimes, nil
}
