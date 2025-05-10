package queries

import (
	"fmt"
	"mc_manager/db"

	"github.com/shopspring/decimal"
)

type mutableConfig struct {
	OfflineImage  *[]byte
	OfflineMOTD   *string
	MCServerPort  int
	IPCheck       bool
	BackupsToKeep int
	TaxRate       decimal.Decimal
}

func GetMutableConfig(DB *db.DB) (*mutableConfig, error) {
	stmt, err := DB.GetStmt("get_config")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow()
	config := &mutableConfig{}
	var taxRateInt int64
	err = result.Scan(&config.OfflineImage, &config.OfflineMOTD, &config.MCServerPort, &config.IPCheck, &config.BackupsToKeep, &taxRateInt)
	if err != nil {
		return nil, err
	}
	if config.MCServerPort == 0 {
		return nil, fmt.Errorf("MCServerPort is not set")
	}

	config.TaxRate = intToDecimal(taxRateInt)
	return config, nil
}

func UpdateOfflineImage(DB *db.DB, image *[]byte) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OfflineImage = image
	return updateConfig(DB, config)
}

func UpdateOfflineMOTD(DB *db.DB, motd *string) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OfflineMOTD = motd
	return updateConfig(DB, config)
}

func UpdateMCServerPort(DB *db.DB, port int) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.MCServerPort = port
	return updateConfig(DB, config)
}

func UpdateIPCheck(DB *db.DB, ipCheck bool) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.IPCheck = ipCheck
	return updateConfig(DB, config)
}

func UpdateBackupsToKeep(DB *db.DB, backupsToKeep int) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.BackupsToKeep = backupsToKeep
	return updateConfig(DB, config)
}

func UpdateTaxRate(DB *db.DB, taxRate decimal.Decimal) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.TaxRate = taxRate
	return updateConfig(DB, config)
}

func updateConfig(DB *db.DB, config *mutableConfig) error {
	stmt, err := DB.GetStmt("update_config")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(config.OfflineImage, config.OfflineMOTD, config.MCServerPort, config.IPCheck, config.BackupsToKeep, decimalToInt(config.TaxRate))
	if err != nil {
		return err
	}
	return nil
}
