package cost_watcher_test

import (
	"math"
	"math/rand/v2"
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"mc_analyser/logging"
	"mc_analyser/model"
	"path"
	"sort"
	"testing"
	"time"
	_ "unsafe"

	"mc_analyser/processes/cost_watcher"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

//go:linkname createBill mc_analyser/processes/cost_watcher.(*CostWatcher).createBill
func createBill(*cost_watcher.CostWatcher, model.YearMonth) error

type mockAWSAPI struct {
	model.AWSAPI
	NextCostResponse  []*model.MonthsCost
	NextErrorResponse error
	Requests          []*model.Timeframe
}

func (m *mockAWSAPI) GetCostForTimeframe(timeframe model.Timeframe) ([]*model.MonthsCost, error) {
	m.Requests = append(m.Requests, &timeframe)
	if m.NextErrorResponse != nil {
		return nil, m.NextErrorResponse
	}
	return m.NextCostResponse, m.NextErrorResponse
}

func TestCreateBill(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	aws := &mockAWSAPI{
		NextCostResponse:  []*model.MonthsCost{},
		NextErrorResponse: nil,
		Requests:          []*model.Timeframe{},
	}
	app := &model.App{
		Db:  DB,
		AWS: aws,
	}
	cw := cost_watcher.NewCostWatcher(app)
	app.CW = cw
	unbilledMonth, err := model.NewYearMonthFromInt(2023, 10)
	assert.NoError(t, err)

	// player is allowed and has playtime
	playerUUID, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3f")
	assert.NoError(t, err)
	invoiceServiceUserId := "testInvoiceServiceUserId"
	err = queries.AddPlayer(DB, *playerUUID, true, true, nil, []string{"Test"}, &invoiceServiceUserId)
	assert.NoError(t, err)

	// player is allowed and has no playtime
	playerUUID2, err := model.NewUUIDFromString("210c22ac-1019-495a-ae72-a9707bfef19d")
	assert.NoError(t, err)
	invoiceServiceUserId2 := "testInvoiceServiceUserId2"
	err = queries.AddPlayer(DB, *playerUUID2, true, true, nil, []string{"Test2"}, &invoiceServiceUserId2)
	assert.NoError(t, err)

	// player is not allowed and has playtime
	playerUUID3, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4c")
	assert.NoError(t, err)
	invoiceServiceUserId3 := "testInvoiceServiceUserId3"
	err = queries.AddPlayer(DB, *playerUUID3, false, true, nil, []string{"Test3"}, &invoiceServiceUserId3)
	assert.NoError(t, err)

	// player is not allowed and has no playtime
	playerUUID4, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4d")
	assert.NoError(t, err)
	invoiceServiceUserId4 := "testInvoiceServiceUserId4"
	err = queries.AddPlayer(DB, *playerUUID4, false, true, nil, []string{"Test4"}, &invoiceServiceUserId4)
	assert.NoError(t, err)

	// another player that is allowed and has playtime
	playerUUID5, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4e")
	assert.NoError(t, err)
	invoiceServiceUserId5 := "testInvoiceServiceUserId5"
	err = queries.AddPlayer(DB, *playerUUID5, true, true, nil, []string{"Test5"}, &invoiceServiceUserId5)
	assert.NoError(t, err)

	// another player that is allowed and has playtime
	playerUUID6, err := model.NewUUIDFromString("6f3b2c4e-5a0d-4f1b-8a7c-9d6e0f2f3b4f")
	assert.NoError(t, err)
	invoiceServiceUserId6 := "testInvoiceServiceUserId6"
	err = queries.AddPlayer(DB, *playerUUID6, true, true, nil, []string{"Test6"}, &invoiceServiceUserId6)
	assert.NoError(t, err)

	players, err := queries.GetPlayers(DB)
	assert.NoError(t, err)
	assert.Equal(t, len(players), 6)

	// add player events
	startDate := unbilledMonth.StartTime().Add(-time.Hour * 24)
	startDate = startDate.Truncate(time.Second)
	playerEvents := make([]*model.PlayerEvent, 0)
	for _, playerId := range []int64{1, 3, 5, 6} {
		currentDate := startDate
		for i := range 200 {
			nextEventDuration := time.Second * time.Duration(math.Min(72000, math.Max(60, math.Abs(math.Ceil(rand.NormFloat64()*3600*5+60)))))
			nextEventDuration = nextEventDuration.Truncate(time.Second)
			if i%2 == 0 {
				nextEventDuration *= 3
			}

			currentDate = currentDate.Add(nextEventDuration)
			err = queries.AddPlayerEvent(DB, currentDate, playerId, i%2 == 0)
			assert.NoError(t, err)
			playerEvents = append(playerEvents, &model.PlayerEvent{
				Timestamp: currentDate,
				PlayerId:  playerId,
				IsOnline:  i%2 == 0,
			})
		}
	}

	// add server events
	playersOnline := 0
	sort.Slice(playerEvents, func(i, j int) bool { return playerEvents[i].Timestamp.Before(playerEvents[j].Timestamp) })
	for _, playerEvent := range playerEvents {
		if playerEvent.IsOnline {
			if playersOnline == 0 {
				err = queries.AddServerEvent(DB, playerEvent.Timestamp.Add(time.Duration(-60)*time.Second), true)
				assert.NoError(t, err)
			}
			playersOnline++
		} else {
			playersOnline--
			if playersOnline == 0 {
				err = queries.AddServerEvent(DB, playerEvent.Timestamp.Add(time.Duration(120)*time.Second), false)
				assert.NoError(t, err)
			}
		}
	}

	cpuTime, err := queries.CalculateCPUTime(DB, unbilledMonth.Timeframe())
	assert.NoError(t, err)
	logging.Debug("CPU time", "cpuTime", cpuTime)

	serverUptime, err := queries.CalculateServerUptime(DB, unbilledMonth.Timeframe())
	assert.NoError(t, err)
	logging.Debug("Server uptime", "uptime", serverUptime)

	fixedCost := decimal.NewFromFloat(0.05).Mul(decimal.NewFromInt(31))
	variableCost := decimal.NewFromFloat(0.0432).Mul(decimal.NewFromFloat(serverUptime.Hours()))
	totalCost := fixedCost.Add(variableCost)

	// Create AWS response
	aws.NextCostResponse = []*model.MonthsCost{
		{
			Timeframe:    unbilledMonth.Timeframe(),
			BillingMonth: unbilledMonth,
			Total:        totalCost,
			Fixed:        fixedCost,
			Variable:     variableCost,
			Currency:     "USD",
			Items: []model.CostItem{
				{
					Label:   "Test Item",
					Amount:  fixedCost,
					IsFixed: true,
				}, {
					Label:   "Test Item 2",
					Amount:  variableCost.Mul(decimal.NewFromFloat(0.8)),
					IsFixed: false,
				}, {
					Label:   "Test Item 3",
					Amount:  variableCost.Mul(decimal.NewFromFloat(0.2)),
					IsFixed: false,
				},
			},
		},
	}
	logging.Debug("cost", "total", totalCost, "fixed", fixedCost, "variable", variableCost)

	// Call the createBill method
	err = createBill(cw, unbilledMonth)
	assert.NoError(t, err)

	//get the bill
	bill, err := queries.GetFullBillByBillingMonth(DB, unbilledMonth)
	assert.NoError(t, err)
	assert.NotNil(t, bill)
	assert.Equal(t, unbilledMonth, bill.Bill.BillingMonth)
	assert.True(t, totalCost.Round(2).Equal(bill.Bill.TotalAmount.Round(2)), "totalCost: %s, bill.Bill.TotalAmount: %s", totalCost.String(), bill.Bill.TotalAmount.String())
	assert.True(t, fixedCost.Round(2).Equal(bill.Bill.FixedAmount.Round(2)), "fixedCost: %s, bill.Bill.FixedAmount: %s", fixedCost.String(), bill.Bill.FixedAmount.String())
	assert.True(t, variableCost.Round(2).Equal(bill.Bill.VariableAmount.Round(2)), "variableCost: %s, bill.Bill.VariableAmount: %s", variableCost.String(), bill.Bill.VariableAmount.String())
	assert.Equal(t, serverUptime, bill.Bill.ServerUptime)
	assert.Equal(t, "USD", bill.Bill.Currency)
	assert.Equal(t, 3, len(bill.BillItems))
	assert.Equal(t, 4, len(bill.BilledPlayers))

	totalFixedCost := decimal.NewFromFloat(0)
	totalVariableCost := decimal.NewFromFloat(0)
	for _, billItem := range bill.BillItems {
		assert.Equal(t, billItem.BillId, bill.Bill.Id)
		if billItem.IsFixed {
			totalFixedCost = totalFixedCost.Add(billItem.Amount)
		} else {
			totalVariableCost = totalVariableCost.Add(billItem.Amount)
		}
	}
	assert.True(t, totalFixedCost.Round(2).Equal(bill.Bill.FixedAmount.Round(2)), "totalFixedCost: %s, bill.Bill.FixedAmount: %s", totalFixedCost.String(), bill.Bill.FixedAmount.String())
	assert.True(t, totalVariableCost.Round(2).Equal(bill.Bill.VariableAmount.Round(2)), "totalVariableCost: %s, bill.Bill.VariableAmount: %s", totalVariableCost.String(), bill.Bill.VariableAmount.String())

	totalFixedCost = decimal.NewFromFloat(0)
	totalVariableCost = decimal.NewFromFloat(0)
	for _, billedPlayer := range bill.BilledPlayers {
		assert.Equal(t, bill.Bill.Id, billedPlayer.BillId)
		totalFixedCost = totalFixedCost.Add(billedPlayer.FixedAmount)
		totalVariableCost = totalVariableCost.Add(billedPlayer.VariableAmount)
	}
	assert.True(t, totalFixedCost.Round(2).Equal(bill.Bill.FixedAmount.Round(2)), "totalFixedCost: %s, bill.Bill.FixedAmount: %s", totalFixedCost.String(), bill.Bill.FixedAmount.String())
	assert.True(t, totalVariableCost.Round(2).Equal(bill.Bill.VariableAmount.Round(2)), "totalVariableCost: %s, bill.Bill.VariableAmount: %s", totalVariableCost.String(), bill.Bill.VariableAmount.String())

	// Check if explored cost are updated
	exploredCost, err := queries.GetExploredCost(DB, unbilledMonth)
	assert.NoError(t, err)
	assert.NotNil(t, exploredCost)
	assert.Equal(t, unbilledMonth, exploredCost.ExploredMonth)

}
