package queries

import (
	"database/sql"
	"mc_manager/db"
	"mc_manager/logging"
	"mc_manager/model"
	"time"

	"github.com/shopspring/decimal"
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

func GetBilledTotalCost(DB *db.DB) (*decimal.Decimal, error) {
	stmt, err := DB.GetStmt("get_billed_total_cost")
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow()
	err = result.Err()
	if err != nil {
		return &decimal.Zero, err
	}

	var billedTotalCostInt int64
	err = result.Scan(&billedTotalCostInt)
	if err == sql.ErrNoRows {
		return &decimal.Zero, nil
	}
	if err != nil {
		return &decimal.Zero, err
	}

	billedTotalCost := intToDecimal(billedTotalCostInt)
	return &billedTotalCost, nil
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

func GetUnsyncedBills(DB *db.DB) ([]*model.Bill, error) {
	stmt, err := DB.GetStmt("get_unsynced_bills")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*model.Bill
	for rows.Next() {
		bill, err := parseBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
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

func GetBillingMonths(DB *db.DB) ([]model.YearMonth, error) {
	stmt, err := DB.GetStmt("get_billing_months")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	var billingMonths []model.YearMonth
	for rows.Next() {
		var billingMonth model.YearMonth
		err = rows.Scan(&billingMonth)
		if err != nil {
			return nil, err
		}
		billingMonths = append(billingMonths, billingMonth)
	}
	return billingMonths, nil
}

func GetBilledPlayerTime(DB *db.DB) ([]*model.BilledPlayerTime, error) {
	stmt, err := DB.GetStmt("get_billed_player_time")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	var billedTimes []*model.BilledPlayerTime
	for rows.Next() {
		var billedTime model.BilledPlayerTime
		err = rows.Scan(&billedTime.PlayerId, &billedTime.Playtime, &billedTime.CPUTime)
		if err != nil {
			return nil, err
		}
		billedTimes = append(billedTimes, &billedTime)
	}
	return billedTimes, nil
}

func GetBilledPlayerTimeForTimespan(DB *db.DB, start int, end int) ([]*model.BilledPlayerTime, error) {
	stmt, err := DB.GetStmt("get_billed_player_time_for_timespan")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(start, end)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	var billedTimes []*model.BilledPlayerTime
	for rows.Next() {
		var billedTime model.BilledPlayerTime
		err = rows.Scan(&billedTime.PlayerId, &billedTime.Playtime, &billedTime.CPUTime)
		if err != nil {
			return nil, err
		}
		billedTimes = append(billedTimes, &billedTime)
	}
	return billedTimes, nil
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
