package queries

import (
	"database/sql"
	"mc_analyser/db"
	"mc_analyser/model"
	"time"
)

func parseServerEvent(row scannable) (*model.ServerEvent, error) {
	var id int64
	var timestamp time.Time
	var isOnline bool
	err := row.Scan(&id, &timestamp, &isOnline)
	if err != nil {
		return nil, err
	}

	return &model.ServerEvent{
		Id:        id,
		Timestamp: timestamp,
		IsOnline:  isOnline,
	}, nil
}

func AddServerEvent(DB *db.DB, timestamp time.Time, isOnline bool) error {
	lastState, err := GetLastServerEvent(DB)
	if err == sql.ErrNoRows {
		lastState = &model.ServerEvent{
			Id:        0,
			Timestamp: time.Now(),
			IsOnline:  false,
		}
	} else if err != nil {
		return err
	}

	if lastState.IsOnline == isOnline {
		return nil
	}

	stmt, err := DB.GetQuery("add_server_event")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(timestamp, isOnline)
	return err
}

func GetLastServerEvent(DB *db.DB) (*model.ServerEvent, error) {
	stmt, err := DB.GetQuery("get_last_server_event")
	if err != nil {
		return &model.ServerEvent{}, err
	}

	result := stmt.QueryRow()
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parseServerEvent(result)
}

func GetServerEvents(DB *db.DB, limit int) ([]*model.ServerEvent, error) {
	stmt, err := DB.GetQuery("get_server_events")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(limit)
	if err != nil {
		return nil, err
	}

	events := []*model.ServerEvent{}
	for result.Next() {
		event, err := parseServerEvent(result)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
