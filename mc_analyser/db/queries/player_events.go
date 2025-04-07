package queries

import (
	"database/sql"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
	"time"
)

func parsePlayerEvent(row scannable) (*model.PlayerEvent, error) {
	var id int
	var timestamp time.Time
	var playerId int
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

func AddPlayerEvent(DB *db.DB, timestamp time.Time, player_id int, isOnline bool) error {
	lastState, err := GetLastPlayerEvent(DB, player_id)
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

	stmt, err := DB.GetQuery("add_player_event")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(timestamp, player_id, isOnline)
	return err
}

func GetLastPlayerEvent(DB *db.DB, playerId int) (*model.PlayerEvent, error) {
	stmt, err := DB.GetQuery("get_last_player_event")
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

func GetLastPlayerEvents(DB *db.DB) ([]*model.PlayerEvent, error) {
	stmt, err := DB.GetQuery("get_last_player_events")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
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

func CalculatePlaytime(DB *db.DB, start time.Time, end time.Time) ([]*model.Playtime, error) {
	testStart := time.Now()
	stmt, err := DB.GetQuery("calculate_playtime")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(start, start, end, start, start, end, end)
	if err != nil {
		return nil, err
	}

	totalPlaytime := []*model.Playtime{}
	for result.Next() {
		var playerId int
		var playtimeInSec int64
		err := result.Scan(&playerId, &playtimeInSec)
		if err != nil {
			return nil, err
		}

		playtime := time.Duration(playtimeInSec) * time.Second
		totalPlaytime = append(totalPlaytime, &model.Playtime{
			PlayerId: int(playerId),
			Playtime: playtime,
		})
	}
	logging.Debug("Playtime calculation took", "duration", time.Since(testStart))
	return totalPlaytime, nil
}
