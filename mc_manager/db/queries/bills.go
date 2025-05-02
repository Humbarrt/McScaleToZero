package queries

import (
	"mc_manager/db"
	logging "mc_managerlogging"
	model "mc_managermodel"
	"time"
)

func parseBill(row scannable) (*model.Bill, error) {
	bill := &model.Bill{}
	var totalAmountInt, fixedAmountInt, variableAmountInt, serverUptimeSeconds int64
	err := row.Scan(&bill.Id,
		&bill.CreatedAt,
		&bill.BillingMonth,
		&totalAmountInt,
		&fixedAmountInt,
		&variableAmountInt,
		&serverUptimeSeconds,
		&bill.Currency,
		&bill.InvoiceServiceInvoiceId)
	if err != nil {
		return nil, err
	}
	bill.TotalAmount = intToDecimal(totalAmountInt)
	bill.FixedAmount = intToDecimal(fixedAmountInt)
	bill.VariableAmount = intToDecimal(variableAmountInt)
	bill.ServerUptime = time.Duration(serverUptimeSeconds) * time.Second
	return bill, nil
}

func parseBillItem(row scannable) (*model.BillItem, error) {
	billItem := &model.BillItem{}
	var amountInt int64
	err := row.Scan(&billItem.Id,
		&billItem.BillId,
		&billItem.Label,
		&billItem.IsFixed,
		&amountInt)
	if err != nil {
		return nil, err
	}

	billItem.Amount = intToDecimal(amountInt)
	return billItem, nil
}

func parseBilledPlayer(row scannable) (*model.BilledPlayer, error) {
	billedPlayer := &model.BilledPlayer{}
	var totalAmountInt, fixedAmountInt, variableAmountInt int64
	err := row.Scan(&billedPlayer.Id,
		&billedPlayer.BillId,
		&billedPlayer.PlayerId,
		&totalAmountInt,
		&fixedAmountInt,
		&variableAmountInt,
		&billedPlayer.Playtime,
		&billedPlayer.CPUTime)
	if err != nil {
		return nil, err
	}

	billedPlayer.TotalAmount = intToDecimal(totalAmountInt)
	billedPlayer.FixedAmount = intToDecimal(fixedAmountInt)
	billedPlayer.VariableAmount = intToDecimal(variableAmountInt)
	return billedPlayer, nil
}

func GetFullBillByBillingMonth(DB *db.DB, billing_month model.YearMonth) (*model.FullBill, error) {
	FullBill := &model.FullBill{}
	bill, err := GetBillByBillingMonth(DB, billing_month)
	if err != nil {
		return nil, err
	}
	FullBill.Bill = bill

	billItems, err := GetBillItemsByBillId(DB, bill.Id)
	if err != nil {
		return nil, err
	}
	FullBill.BillItems = billItems

	billedPlayers, err := GetBilledPlayersByBillId(DB, bill.Id)
	if err != nil {
		return nil, err
	}
	FullBill.BilledPlayers = billedPlayers
	return FullBill, nil
}

func GetBillById(DB *db.DB, billId int64) (*model.Bill, error) {
	stmt, err := DB.GetStmt("get_bill_by_id")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(billId)
	return parseBill(row)
}

func GetBillByBillingMonth(DB *db.DB, billing_month model.YearMonth) (*model.Bill, error) {
	stmt, err := DB.GetStmt("get_bill_by_billing_month")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(billing_month)
	return parseBill(row)
}

func GetLatestBill(DB *db.DB) (*model.Bill, error) {
	stmt, err := DB.GetStmt("get_latest_bill")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow()
	return parseBill(row)
}

func GetLatestBillWithFixedCost(DB *db.DB) (*model.Bill, error) {
	stmt, err := DB.GetStmt("get_latest_bill_with_fixed_cost")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow()
	return parseBill(row)
}

func GetLatestBillWithVariableCost(DB *db.DB) (*model.Bill, error) {
	stmt, err := DB.GetStmt("get_latest_bill_with_variable_cost")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow()
	return parseBill(row)
}

func GetBillItemsByBillId(DB *db.DB, billId int64) ([]*model.BillItem, error) {
	stmt, err := DB.GetStmt("get_bill_items_by_bill_id")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(billId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var billItems []*model.BillItem
	for rows.Next() {
		billItem, err := parseBillItem(rows)
		if err != nil {
			return nil, err
		}
		billItems = append(billItems, billItem)
	}
	return billItems, nil
}

func GetBilledPlayersByBillId(DB *db.DB, billId int64) ([]*model.BilledPlayer, error) {
	stmt, err := DB.GetStmt("get_billed_players_by_bill_id")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(billId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var billedPlayers []*model.BilledPlayer
	for rows.Next() {
		billedPlayer, err := parseBilledPlayer(rows)
		if err != nil {
			return nil, err
		}
		billedPlayers = append(billedPlayers, billedPlayer)
	}
	return billedPlayers, nil
}

func CreateBill(DB *db.DB, bill *model.Bill, billItems []*model.BillItem, billedPlayers []*model.BilledPlayer) error {
	tx, err := DB.NewTx()
	if err != nil {
		return err
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

	addBillStmt, err := tx.GetStmt("add_bill")
	if err != nil {
		return err
	}
	addBillItemStmt, err := tx.GetStmt("add_bill_item")
	if err != nil {
		return err
	}
	addBilledPlayerStmt, err := tx.GetStmt("add_billed_player")
	if err != nil {
		return err
	}

	billRes, err := addBillStmt.Exec(bill.BillingMonth, decimalToInt(bill.TotalAmount), decimalToInt(bill.FixedAmount), decimalToInt(bill.VariableAmount), int64(bill.ServerUptime.Seconds()), bill.Currency, bill.InvoiceServiceInvoiceId)
	if err != nil {
		return err
	}

	billId, err := billRes.LastInsertId()
	if err != nil {
		return err
	}

	for _, item := range billItems {
		_, err = addBillItemStmt.Exec(billId, item.Label, item.IsFixed, decimalToInt(item.Amount))
		if err != nil {
			return err
		}
	}

	for _, player := range billedPlayers {
		_, err = addBilledPlayerStmt.Exec(billId, player.PlayerId, decimalToInt(player.TotalAmount), decimalToInt(player.FixedAmount), decimalToInt(player.VariableAmount), player.Playtime, player.CPUTime)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		logging.Error("Error committing transaction", "error", err)
		return err
	}
	return nil
}

func AddInvoiceServiceInvoiceId(DB *db.DB, billId int64, invoiceServiceInvoiceId string) error {
	stmt, err := DB.GetStmt("add_invoice_service_invoice_id")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(invoiceServiceInvoiceId, billId)
	if err != nil {
		return err
	}
	return nil
}
