package queries

import (
	"mc_manager/db"
)

func ResetInvoiceService(DB *db.DB) error {
	stmt, err := DB.GetStmt("reset_invoice_service")
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	return err
}
