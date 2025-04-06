package queries

import (
	"fmt"
	"mc_analyser/db"
)

type dynamicConfig struct {
	OfflineImage *[]byte
	OfflineMOTD  *string
	MCServerPort int
}

func GetConfig(DB *db.DB) (*dynamicConfig, error) {
	stmt, err := DB.GetQuery("get_config")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow()
	config := &dynamicConfig{}
	err = result.Scan(&config.OfflineImage, &config.OfflineMOTD, &config.MCServerPort)
	if err != nil {
		return nil, err
	}
	if config.MCServerPort == 0 {
		return nil, fmt.Errorf("MCServerPort is not set")
	}

	return config, nil
}

func UpdateOfflineImage(DB *db.DB, image *[]byte) error {
	stmt, err := DB.GetQuery("update_offline_image")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(image)
	if err != nil {
		return err
	}
	return nil
}

func UpdateOfflineMOTD(DB *db.DB, motd *string) error {
	stmt, err := DB.GetQuery("update_offline_motd")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(motd)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMCServerPort(DB *db.DB, port int) error {
	stmt, err := DB.GetQuery("update_mc_server_port")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(port)
	if err != nil {
		return err
	}
	return nil
}
