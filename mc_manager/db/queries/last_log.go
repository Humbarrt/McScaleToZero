package queries

import (
	"fmt"
	"mc_manager/db"
	"mc_manager/model"
)

func GetLastLog(DB *db.DB) (*model.Log, error) {
	stmt, err := DB.GetStmt("get_last_log")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow()
	var lastLog model.Log
	var id int
	err = row.Scan(&id, &lastLog.Timestamp, &lastLog.Message)
	if err != nil {
		return nil, err
	}
	if id != 1 {
		return nil, fmt.Errorf("last log id is not 1")
	}

	return &lastLog, nil
}

func SetLastLog(DB *db.DB, lastLog *model.Log) error {
	stmt, err := DB.GetStmt("set_last_log")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(lastLog.Timestamp, lastLog.Message)
	if err != nil {
		return err
	}
	return nil
}
