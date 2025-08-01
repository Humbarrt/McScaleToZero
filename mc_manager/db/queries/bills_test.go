package queries_test

import (
	"database/sql"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/model"
	"path"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBills(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	playerUUID, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	invoiceServiceUserId := "testInvoiceServiceUserId"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID, true, true, nil, []string{"Test"}, &invoiceServiceUserId)
	assert.NoError(t, err)

	player, err := queries.GetPlayerByAlias(DB, "Test")
	assert.NoError(t, err)
	assert.Equal(t, player.McUUID.Bytes(), playerUUID.Bytes())

	playerUUID2, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	invoiceServiceUserId2 := "testInvoiceServiceUserId2"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID2, true, true, nil, []string{"Test2"}, &invoiceServiceUserId2)
	assert.NoError(t, err)

	player2, err := queries.GetPlayerByAlias(DB, "Test2")
	assert.NoError(t, err)
	assert.Equal(t, player2.McUUID.Bytes(), playerUUID2.Bytes())

	playerUUID3, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4c")
	invoiceServiceUserId3 := "testInvoiceServiceUserId3"
	assert.NoError(t, err)
	err = queries.AddPlayer(DB, *playerUUID3, true, true, nil, []string{"Test3"}, &invoiceServiceUserId3)
	assert.NoError(t, err)

	player3, err := queries.GetPlayerByAlias(DB, "Test3")
	assert.NoError(t, err)
	assert.Equal(t, player3.McUUID.Bytes(), playerUUID3.Bytes())

	ym, err := model.NewYearMonthFromInt(2023, 10)
	assert.NoError(t, err)
	bill := model.Bill{
		BillingMonth:            ym,
		TotalAmount:             decimal.NewFromInt(100),
		FixedAmount:             decimal.NewFromInt(50),
		VariableAmount:          decimal.NewFromInt(50),
		ServerUptime:            time.Duration(1240) * time.Hour,
		Currency:                "USD",
		InvoiceServiceInvoiceId: nil,
	}
	billItems := []*model.BillItem{
		{
			Amount:  decimal.NewFromInt(50),
			Label:   "Fixed Test item",
			BillId:  0,
			IsFixed: true,
		},
		{
			Amount:  decimal.NewFromInt(50),
			Label:   "Fixed Test item",
			BillId:  0,
			IsFixed: false,
		},
	}
	billedPlayers := []*model.BilledPlayer{
		{
			PlayerId:       player.Id,
			TotalAmount:    decimal.NewFromInt(50),
			FixedAmount:    decimal.NewFromInt(25),
			VariableAmount: decimal.NewFromInt(25),
			Playtime:       time.Duration(12) * time.Hour,
			CPUTime:        time.Duration(6) * time.Hour,
		},
		{
			PlayerId:       player2.Id,
			TotalAmount:    decimal.NewFromInt(50),
			FixedAmount:    decimal.NewFromInt(25),
			VariableAmount: decimal.NewFromInt(25),
			Playtime:       time.Duration(10) * time.Hour,
			CPUTime:        time.Duration(4) * time.Hour,
		},
		{
			PlayerId:       player3.Id,
			TotalAmount:    decimal.NewFromInt(-50),
			FixedAmount:    decimal.NewFromInt(25),
			VariableAmount: decimal.NewFromInt(25),
			Playtime:       time.Duration(5) * time.Hour,
			CPUTime:        time.Duration(2) * time.Hour,
		},
	}

	// Test if CreateBill tx rollback works
	err = queries.CreateBill(DB, &bill, billItems, billedPlayers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")

	billRes, err := queries.GetLatestBill(DB)
	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Nil(t, billRes)

	// Test CreateBill
	billedPlayers[2].TotalAmount = decimal.NewFromInt(50)
	err = queries.CreateBill(DB, &bill, billItems, billedPlayers)
	assert.NoError(t, err)

	billRes, err = queries.GetLatestBill(DB)
	assert.NoError(t, err)
	assert.Equal(t, bill.BillingMonth, billRes.BillingMonth)
	assert.True(t, bill.TotalAmount.Equal(billRes.TotalAmount))
	assert.True(t, bill.FixedAmount.Equal(billRes.FixedAmount))
	assert.True(t, bill.VariableAmount.Equal(billRes.VariableAmount))
	assert.Equal(t, bill.ServerUptime, billRes.ServerUptime)
	assert.Equal(t, bill.Currency, billRes.Currency)
	assert.Equal(t, bill.InvoiceServiceInvoiceId, billRes.InvoiceServiceInvoiceId)

	// Test GetFullBillByBillingMonth -- includes (GetBillByBillingMonth, GetBillItemsByBillId, GetBilledPlayersByBillId)
	fullBill, err := queries.GetFullBillByBillingMonth(DB, bill.BillingMonth)
	assert.NoError(t, err)
	assert.Equal(t, bill.BillingMonth, fullBill.Bill.BillingMonth)
	assert.True(t, bill.TotalAmount.Equal(fullBill.Bill.TotalAmount))
	assert.True(t, bill.FixedAmount.Equal(fullBill.Bill.FixedAmount))
	assert.True(t, bill.VariableAmount.Equal(fullBill.Bill.VariableAmount))
	assert.Equal(t, bill.ServerUptime, fullBill.Bill.ServerUptime)
	assert.Equal(t, bill.Currency, fullBill.Bill.Currency)
	assert.Equal(t, bill.InvoiceServiceInvoiceId, fullBill.Bill.InvoiceServiceInvoiceId)

	assert.Len(t, fullBill.BillItems, len(billItems))
	for i, billItem := range fullBill.BillItems {
		assert.Equal(t, billItems[i].Label, billItem.Label)
		assert.Equal(t, billItems[i].IsFixed, billItem.IsFixed)
		assert.True(t, billItems[i].Amount.Equal(billItem.Amount))
		assert.Equal(t, fullBill.Bill.Id, billItem.BillId)
	}

	assert.Len(t, fullBill.BilledPlayers, len(billedPlayers))
	for i, billedPlayer := range fullBill.BilledPlayers {
		assert.Equal(t, billedPlayers[i].PlayerId, billedPlayer.PlayerId)
		assert.True(t, billedPlayers[i].TotalAmount.Equal(billedPlayer.TotalAmount))
		assert.True(t, billedPlayers[i].FixedAmount.Equal(billedPlayer.FixedAmount))
		assert.True(t, billedPlayers[i].VariableAmount.Equal(billedPlayer.VariableAmount))
		assert.Equal(t, billedPlayers[i].Playtime, billedPlayer.Playtime)
		assert.Equal(t, billedPlayers[i].CPUTime, billedPlayer.CPUTime)
		assert.Equal(t, fullBill.Bill.Id, billedPlayer.BillId)
	}

	// Test GetLatestBill
	bill.BillingMonth = model.NewYearMonthFromTime(time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC))
	err = queries.CreateBill(DB, &bill, billItems, billedPlayers)
	assert.NoError(t, err)
	billRes, err = queries.GetLatestBill(DB)
	assert.NoError(t, err)
	assert.Equal(t, bill.BillingMonth, billRes.BillingMonth)

	// Test GetBillByBillId
	billRes, err = queries.GetBillById(DB, billRes.Id)
	assert.NoError(t, err)
	assert.Equal(t, billRes.Id, billRes.Id)

	// Test GetBillByBillingMonth
	billRes, err = queries.GetBillByBillingMonth(DB, bill.BillingMonth)
	assert.NoError(t, err)
	assert.Equal(t, billRes.Id, billRes.Id)

	// Test duplicate bill with rollback
	err = queries.CreateBill(DB, &bill, billItems, billedPlayers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed: bills.billing_month")

	_billRes, err := queries.GetBillById(DB, billRes.Id+1)
	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Nil(t, _billRes)

	// Test adding invoice service invoice id
	err = queries.AddInvoiceServiceInvoiceId(DB, billRes.Id, "testInvoiceServiceInvoiceId")
	assert.NoError(t, err)
	queriedId := billRes.Id
	billRes, err = queries.GetBillById(DB, queriedId)
	assert.NoError(t, err)
	assert.NotNil(t, billRes.InvoiceServiceInvoiceId)
	assert.Equal(t, *billRes.InvoiceServiceInvoiceId, "testInvoiceServiceInvoiceId")
	assert.Equal(t, queriedId, billRes.Id)

	bill = model.Bill{
		BillingMonth:            ym.AddMonths(2),
		TotalAmount:             decimal.NewFromInt(100),
		FixedAmount:             decimal.NewFromInt(0),
		VariableAmount:          decimal.NewFromInt(0),
		ServerUptime:            time.Duration(1240) * time.Hour,
		Currency:                "USD",
		InvoiceServiceInvoiceId: nil,
	}

	err = queries.CreateBill(DB, &bill, billItems, billedPlayers)
	assert.NoError(t, err)
	latestBill, err := queries.GetLatestBill(DB)
	assert.NoError(t, err)

	// Test GetLatestBillWithFixedCost
	notLatestBill, err := queries.GetLatestBillWithFixedCost(DB)
	assert.NoError(t, err)
	assert.Equal(t, billRes.Id, notLatestBill.Id)
	assert.NotEqual(t, latestBill.Id, notLatestBill.Id)

	// Test GetLatestBillWithVariableCost
	notLatestBill, err = queries.GetLatestBillWithVariableCost(DB)
	assert.NoError(t, err)
	assert.Equal(t, billRes.Id, notLatestBill.Id)
	assert.NotEqual(t, latestBill.Id, notLatestBill.Id)

	// Test GetBillingMonths
	billingMonths, err := queries.GetBillingMonths(DB)
	assert.NoError(t, err)
	assert.Len(t, billingMonths, 3)
	assert.Contains(t, billingMonths, ym)
}
