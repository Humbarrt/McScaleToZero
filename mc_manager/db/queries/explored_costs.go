package queries

import (
	"mc_manager/db"
	model "mc_managermodel"
)

func GetExploredCost(DB *db.DB, month model.YearMonth) (*model.ExploredCost, error) {
	stmt, err := DB.GetStmt("get_explored_cost")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(month)
	var cost model.ExploredCost
	err = row.Scan(&cost.Id, &cost.CreatedAt, &cost.ExploredMonth)
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

func AddExploredCost(DB *db.DB, month model.YearMonth) error {
	stmt, err := DB.GetStmt("add_explored_cost")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(month)
	if err != nil {
		return err
	}
	return nil
}
