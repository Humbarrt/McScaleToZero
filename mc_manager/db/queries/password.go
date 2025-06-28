package queries

import (
	"database/sql"
	"mc_manager/db"
)

func GetClientKey(DB *db.DB, playerId int64) ([]byte, error) {
	stmt, err := DB.GetStmt("get_client_key")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(playerId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	var key *[]byte
	err = result.Scan(&key)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, sql.ErrNoRows
	}

	return *key, nil
}

func UpdatePlayerPassword(DB *db.DB, playerId int64, encryptedPassword []byte, clientKey []byte) error {
	stmt, err := DB.GetStmt("update_player_password")
	if err != nil {
		return err
	}

	result, err := stmt.Exec(encryptedPassword, clientKey, playerId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func GetPlayerPassword(DB *db.DB, playerId int64) ([]byte, error) {
	stmt, err := DB.GetStmt("get_player_password")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(playerId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	var password *[]byte
	err = result.Scan(&password)
	if err != nil {
		return nil, err
	}
	if password == nil {
		return nil, sql.ErrNoRows
	}

	return *password, nil
}
