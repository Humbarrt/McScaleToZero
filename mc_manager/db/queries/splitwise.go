package queries

import (
	"encoding/json"
	"fmt"
	"mc_manager/db"
	"mc_manager/model"
)

func GetSplitwiseConfig(DB *db.DB) (*model.SplitwiseConfig, error) {
	stmt, err := DB.GetStmt("get_splitwise_config")
	if err != nil {
		return &model.SplitwiseConfig{}, err
	}

	row := stmt.QueryRow()
	var config model.SplitwiseConfig
	var currenciesString, usersString string
	err = row.Scan(&config.InvoiceCurrency, &currenciesString, &usersString)
	if err != nil {
		return &model.SplitwiseConfig{}, err
	}

	if err := json.Unmarshal([]byte(currenciesString), &config.Currencies); err != nil {
		return &model.SplitwiseConfig{}, fmt.Errorf("error unmarshalling splitwise currencies: %w", err)
	}

	if err := json.Unmarshal([]byte(usersString), &config.Users); err != nil {
		return &model.SplitwiseConfig{}, fmt.Errorf("error unmarshalling splitwise users: %w", err)
	}

	return &config, nil
}

func UpdateSplitwiseConfigInvoiceCurrency(DB *db.DB, invoice_currency string) error {
	splitwiseConfig, err := GetSplitwiseConfig(DB)
	if err != nil {
		return err
	}
	splitwiseConfig.InvoiceCurrency = invoice_currency

	return updateSplitwiseConfig(DB, splitwiseConfig)
}

func UpdateSplitwiseConfigCurrencies(DB *db.DB, currencies []string) error {
	splitwiseConfig, err := GetSplitwiseConfig(DB)
	if err != nil {
		return err
	}
	splitwiseConfig.Currencies = currencies

	return updateSplitwiseConfig(DB, splitwiseConfig)
}

func UpdateSplitwiseConfigUsers(DB *db.DB, users []*model.SplitwiseUser) error {
	splitwiseConfig, err := GetSplitwiseConfig(DB)
	if err != nil {
		return err
	}
	splitwiseConfig.Users = users

	return updateSplitwiseConfig(DB, splitwiseConfig)
}

func updateSplitwiseConfig(DB *db.DB, splitwiseConfig *model.SplitwiseConfig) error {
	stmt, err := DB.GetStmt("update_splitwise_config")
	if err != nil {
		return err
	}

	currenciesString, err := json.Marshal(splitwiseConfig.Currencies)
	if err != nil {
		return fmt.Errorf("error marshalling splitwise currencies: %w", err)
	}

	usersString, err := json.Marshal(splitwiseConfig.Users)
	if err != nil {
		return fmt.Errorf("error marshalling splitwise users: %w", err)
	}

	_, err = stmt.Exec(splitwiseConfig.InvoiceCurrency, currenciesString, usersString)
	if err != nil {
		return err
	}
	return nil
}
