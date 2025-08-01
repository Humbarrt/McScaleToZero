package queries

import (
	"fmt"
	"mc_manager/db"
	"mc_manager/model"

	"github.com/shopspring/decimal"
)

func GetMutableConfig(DB *db.DB) (*model.MutableConfig, error) {
	stmt, err := DB.GetStmt("get_config")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow()
	config := &model.MutableConfig{}
	var taxRateInt int64
	var networkRatioInt int64
	var fallbackFixedCostInt int64
	var fallbackVariableCostInt int64
	err = result.Scan(&config.MCServerPort, &config.OnlineImage, &config.OnlineMOTD, &config.OfflineImage, &config.OfflineMOTD, &config.ReplaceMOTD, &config.LastImage, &config.LastMOTD, &config.ProtocolVersion, &config.MaxPlayers, &config.BackupsToKeep, &taxRateInt, &networkRatioInt, &fallbackFixedCostInt, &fallbackVariableCostInt)
	if err != nil {
		return nil, err
	}
	if config.MCServerPort == 0 {
		return nil, fmt.Errorf("MCServerPort is not set")
	}

	config.TaxRate = intToDecimal(taxRateInt)
	config.NetworkRatio = intToDecimal(networkRatioInt)
	config.FallbackFixedCost = intToDecimal(fallbackFixedCostInt)
	config.FallbackVariableCost = intToDecimal(fallbackVariableCostInt)
	return config, nil
}

func UpdateMCServerPort(DB *db.DB, port int) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.MCServerPort = port
	return updateConfig(DB, config)
}

func UpdateOnlineImage(DB *db.DB, image *string) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OnlineImage = image
	return updateConfig(DB, config)
}

func UpdateOnlineMOTD(DB *db.DB, motd []byte) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OnlineMOTD = motd
	return updateConfig(DB, config)
}

func UpdateOfflineImage(DB *db.DB, image *string) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OfflineImage = image
	return updateConfig(DB, config)
}

func UpdateOfflineMOTD(DB *db.DB, motd []byte) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.OfflineMOTD = motd
	return updateConfig(DB, config)
}

func UpdateReplaceMOTD(DB *db.DB, replaceMOTD bool) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.ReplaceMOTD = replaceMOTD
	return updateConfig(DB, config)
}

func UpdateLastImage(DB *db.DB, image *string) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.LastImage = image
	return updateConfig(DB, config)
}

func UpdateLastMOTD(DB *db.DB, motd []byte) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.LastMOTD = motd
	return updateConfig(DB, config)
}

func UpdateProtocolVersion(DB *db.DB, protocolVersion *[]byte) error {
	if protocolVersion == nil {
		return fmt.Errorf("error protocol version is nil")
	}
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.ProtocolVersion = protocolVersion
	return updateConfig(DB, config)
}

func UpdateMaxPlayers(DB *db.DB, maxPlayers int) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.MaxPlayers = maxPlayers
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

func UpdateNetworkRatio(DB *db.DB, networkRatio decimal.Decimal) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.NetworkRatio = networkRatio
	return updateConfig(DB, config)
}

func UpdateFallbackFixedCost(DB *db.DB, fallbackFixedCost decimal.Decimal) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.FallbackFixedCost = fallbackFixedCost
	return updateConfig(DB, config)
}

func UpdateFallbackVariableCost(DB *db.DB, fallbackVariableCost decimal.Decimal) error {
	config, err := GetMutableConfig(DB)
	if err != nil {
		return err
	}
	config.FallbackVariableCost = fallbackVariableCost
	return updateConfig(DB, config)
}

func updateConfig(DB *db.DB, config *model.MutableConfig) error {
	stmt, err := DB.GetStmt("update_config")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(config.MCServerPort, config.OnlineImage, config.OnlineMOTD, config.OfflineImage, config.OfflineMOTD, config.ReplaceMOTD, config.LastImage, config.LastMOTD, config.ProtocolVersion, config.MaxPlayers, config.BackupsToKeep, decimalToInt(config.TaxRate), decimalToInt(config.NetworkRatio), decimalToInt(config.FallbackFixedCost), decimalToInt(config.FallbackVariableCost))
	if err != nil {
		return err
	}
	return nil
}
