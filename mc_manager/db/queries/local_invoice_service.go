package queries

import (
	"mc_manager/db"
	"mc_manager/logging"
	"mc_manager/model"
	"time"

	"github.com/shopspring/decimal"
)

func GetLocalInvoiceServiceConfig(DB *db.DB) (*model.LocalInvoiceServiceConfig, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_config")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow()
	err = result.Err()
	if err != nil {
		return nil, err
	}

	var invoiceCurrency string
	var payingPlayerId *int64
	err = result.Scan(&invoiceCurrency, &payingPlayerId)
	if err != nil {
		return nil, err
	}

	config := &model.LocalInvoiceServiceConfig{
		InvoiceCurrency: invoiceCurrency,
		PayingPlayerId:  payingPlayerId,
	}

	return config, nil
}

func UpdateLocalInvoiceServiceConfig(DB *db.DB, config *model.LocalInvoiceServiceConfig) error {
	stmt, err := DB.GetStmt("update_local_invoice_service_config")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(config.InvoiceCurrency, config.PayingPlayerId)
	return err
}

func parseLocalInvoicePayment(row scannable) (*model.LocalInvoiceServicePayment, error) {
	var id int64
	var playerId int64
	var amountInt int64
	var description string
	var editedAt *time.Time
	var createdAt time.Time
	err := row.Scan(&id, &playerId, &amountInt, &description, &editedAt, &createdAt)
	if err != nil {
		return nil, err
	}

	return &model.LocalInvoiceServicePayment{
		Id:          id,
		PlayerId:    playerId,
		Amount:      intToDecimal(amountInt),
		Description: description,
		EditedAt:    editedAt,
		CreatedAt:   createdAt,
	}, nil
}

func GetLocalInvoiceServicePayments(DB *db.DB) ([]*model.LocalInvoiceServicePayment, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_payments")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	payments := []*model.LocalInvoiceServicePayment{}
	for result.Next() {
		payment, err := parseLocalInvoicePayment(result)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func GetLocalInvoiceServicePaymentById(DB *db.DB, paymentId int64) (*model.LocalInvoiceServicePayment, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_payment_by_id")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(paymentId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parseLocalInvoicePayment(result)
}

func GetLocalInvoiceServicePaymentsByPlayerId(DB *db.DB, playerId int64) ([]*model.LocalInvoiceServicePayment, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_payments_by_player_id")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(playerId)
	if err != nil {
		return nil, err
	}

	payments := []*model.LocalInvoiceServicePayment{}
	for result.Next() {
		payment, err := parseLocalInvoicePayment(result)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func AddLocalInvoiceServicePayment(DB *db.DB, playerId int64, amount decimal.Decimal, description string) error {
	stmt, err := DB.GetStmt("add_local_invoice_service_payment")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(playerId, decimalToInt(amount), description)
	return err
}

func UpdateLocalInvoiceServicePayment(DB *db.DB, payment *model.LocalInvoiceServicePayment) error {
	stmt, err := DB.GetStmt("update_local_invoice_service_payment")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(payment.PlayerId, decimalToInt(payment.Amount), payment.Description, payment.Id)
	return nil
}

func parseLocalInvoice(row scannable) (*model.LocalInvoiceServiceInvoice, error) {
	var id int64
	var amountInt int64
	var description string
	var createdAt time.Time
	err := row.Scan(&id, &amountInt, &description, &createdAt)
	if err != nil {
		return nil, err
	}

	return &model.LocalInvoiceServiceInvoice{
		Id:          id,
		Amount:      intToDecimal(amountInt),
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

func GetLocalInvoiceServiceInvoices(DB *db.DB) ([]*model.LocalInvoiceServiceInvoice, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_invoices")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	invoices := []*model.LocalInvoiceServiceInvoice{}
	for result.Next() {
		invoice, err := parseLocalInvoice(result)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func GetLocalInvoiceServiceInvoiceById(DB *db.DB, invoiceId int64) (*model.LocalInvoiceServiceInvoice, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_invoice_by_id")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(invoiceId)
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return parseLocalInvoice(result)
}

func AddLocalInvoiceServiceInvoice(DB *db.DB, amount decimal.Decimal, description string, invoiceItems []*model.LocalInvoiceServiceInvoiceItem) (int64, error) {
	tx, err := DB.NewTx()
	if err != nil {
		return -1, err
	}

	//rollback and close transaction if there is an error
	defer func() {
		if err != nil {
			_err := tx.Rollback()
			if _err != nil {
				logging.Error("Error rolling back transaction", "error", _err)
			}
			return
		}
	}()

	addInvoiceStmt, err := tx.GetStmt("add_local_invoice_service_invoice")
	if err != nil {
		return -1, err
	}
	addInvoiceItemStmt, err := tx.GetStmt("add_local_invoice_service_invoice_item")
	if err != nil {
		return -1, err
	}

	invoiceRes, err := addInvoiceStmt.Exec(decimalToInt(amount), description)
	if err != nil {
		return -1, err
	}

	invoiceId, err := invoiceRes.LastInsertId()
	if err != nil {
		return -1, err
	}

	for _, item := range invoiceItems {
		_, err = addInvoiceItemStmt.Exec(invoiceId, item.PlayerId, decimalToInt(item.Amount))
		if err != nil {
			return -1, err
		}
	}

	err = tx.Commit()
	if err != nil {
		logging.Error("Error committing transaction", "error", err)
		return -1, err
	}
	return invoiceId, nil
}

func parseLocalInvoiceItem(row scannable) (*model.LocalInvoiceServiceInvoiceItem, error) {
	var id int64
	var invoiceId int64
	var playerId int64
	var amountInt int64
	err := row.Scan(&id, &invoiceId, &playerId, &amountInt)
	if err != nil {
		return nil, err
	}

	return &model.LocalInvoiceServiceInvoiceItem{
		Id:        id,
		InvoiceId: invoiceId,
		PlayerId:  playerId,
		Amount:    intToDecimal(amountInt),
	}, nil
}

func GetLocalInvoiceServiceInvoiceItemsByInvoiceId(DB *db.DB, invoiceId int64) ([]*model.LocalInvoiceServiceInvoiceItem, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_invoice_items_by_invoice_id")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(invoiceId)
	if err != nil || result.Err() != nil {
		return nil, err
	}

	invoiceItems := []*model.LocalInvoiceServiceInvoiceItem{}
	for result.Next() {
		invoiceItem, err := parseLocalInvoiceItem(result)
		if err != nil {
			return nil, err
		}
		invoiceItems = append(invoiceItems, invoiceItem)
	}

	return invoiceItems, nil
}

func GetLocalInvoiceServiceInvoiceCount(DB *db.DB) (int, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_invoice_count")
	if err != nil {
		return 0, err
	}

	result := stmt.QueryRow()
	err = result.Err()
	if err != nil {
		return 0, err
	}

	var totalInvoices int
	err = result.Scan(&totalInvoices)
	if err != nil {
		return 0, err
	}

	return totalInvoices, nil
}

func CalculateLocalInvoiceServicePlayerCache(DB *db.DB) (map[int64]*model.LocalInvoiceServicePlayerCache, error) {
	stmt, err := DB.GetStmt("calculate_local_invoice_service_player_cache")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
	if err != nil || result.Err() != nil {
		return nil, err
	}

	playerCache := make(map[int64]*model.LocalInvoiceServicePlayerCache)
	for result.Next() {
		var playerId int64
		var totalPaymentsCount int
		var totalPaidAmountInt int64
		var totalInvoicedAmountInt int64

		err := result.Scan(&playerId, &totalPaymentsCount, &totalPaidAmountInt, &totalInvoicedAmountInt)
		if err != nil {
			return nil, err
		}

		totalSpent := intToDecimal(totalInvoicedAmountInt)
		totalPaid := intToDecimal(totalPaidAmountInt)
		outstandingDebt := totalSpent.Sub(totalPaid)
		playerCache[playerId] = &model.LocalInvoiceServicePlayerCache{
			TotalSpent:      totalSpent,
			TotalPaid:       totalPaid,
			OutstandingDebt: outstandingDebt,
			PaymentCount:    totalPaymentsCount,
		}
	}

	return playerCache, nil
}

func GetLocalInvoiceServiceInvoiceItemsByPlayerId(DB *db.DB, playerId int64) ([]*model.ExtendedLocalInvoiceServiceInvoiceItem, error) {
	stmt, err := DB.GetStmt("get_local_invoice_service_invoice_items_by_player_id")
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query(playerId)
	if err != nil || result.Err() != nil {
		return nil, err
	}

	invoiceItems := []*model.ExtendedLocalInvoiceServiceInvoiceItem{}
	for result.Next() {
		var id int64
		var amountInt int64
		var description string
		var createdAt time.Time
		err := result.Scan(&id, &amountInt, &description, &createdAt)
		if err != nil {
			return nil, err
		}

		invoiceItems = append(invoiceItems, &model.ExtendedLocalInvoiceServiceInvoiceItem{
			Id:          id,
			Amount:      intToDecimal(amountInt),
			Description: description,
			CreatedAt:   createdAt,
		})
	}

	return invoiceItems, nil
}
