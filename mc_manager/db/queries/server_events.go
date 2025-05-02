package queries

import (
	"database/sql"
	"mc_manager/db"
	model "mc_managermodel"
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
	lastState, err := GetLatestServerEvent(DB)
	if err == sql.ErrNoRows {
		lastState = &model.ServerEvent{
			Id:        0,
			Timestamp: time.Time{},
			IsOnline:  false,
		}
	} else if err != nil {
		return err
	}

	if lastState.IsOnline == isOnline || !lastState.Timestamp.IsZero() && timestamp.Before(lastState.Timestamp) {
		return nil
	}

	stmt, err := DB.GetStmt("add_server_event")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(timestamp, isOnline)
	return err
}

func GetLatestServerEvent(DB *db.DB) (*model.ServerEvent, error) {
	stmt, err := DB.GetStmt("get_latest_server_event")
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
	stmt, err := DB.GetStmt("get_server_events")
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

func CalculateServerUptime(DB *db.DB, timeframe model.Timeframe) (time.Duration, error) {
	stmt, err := DB.GetStmt("calculate_server_uptime")
	if err != nil {
		return 0, err
	}

	result := stmt.QueryRow(
		sql.Named("start_time", timeframe.Start),
		sql.Named("end_time", timeframe.End),
	)
	err = result.Err()
	if err != nil {
		return 0, err
	}

	var uptime int
	err = result.Scan(&uptime)
	if err != nil {
		return 0, err
	}
	if err == sql.ErrNoRows {
		return 0, nil
	}

	return time.Duration(uptime) * time.Second, nil
}
