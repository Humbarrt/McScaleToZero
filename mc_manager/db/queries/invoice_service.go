package queries

import (
	"mc_manager/db"
)

func ResetInvoiceService(DB *db.DB) error {
	return DB.ExecScript("reset_invoice_service")
}
