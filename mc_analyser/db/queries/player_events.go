package queries

import (
	"database/sql"
	"mc_analyser/db"
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
			Timestamp: time.Now(),
			IsOnline:  false,
		}
	} else if err != nil {
		return err
	}

	if lastState.IsOnline == isOnline {
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
