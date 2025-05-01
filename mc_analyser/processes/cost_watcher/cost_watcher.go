package cost_watcher

import (
	"database/sql"
	"fmt"
	"mc_analyser/db/queries"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type CostWatcher struct {
	App *model.App

	//protected
	mutex           sync.Mutex
	_isWatching     bool
	_lastCostUpdate time.Time
}

func NewCostWatcher(app *model.App) *CostWatcher {
	return &CostWatcher{
		App: app,
	}
}

func (cs *CostWatcher) LastCostUpdate() time.Time {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	return cs._lastCostUpdate
}

func (cs *CostWatcher) setupCostAllocationTag() error {
	for {
		var isTagActive bool
		isBackfilling, err := cs.App.AWS.IsBackfillInProgress()
		if err != nil {
			logging.Error("Error checking backfill status", "error", err)
			goto waitAndContinue
		}
		if isBackfilling {
			logging.Info("Backfill is in progress")
			goto waitAndContinue
		}

		isTagActive, err = cs.App.AWS.IsCostAllocationTagActive()
		if err != nil {
			logging.Error("Error checking cost allocation tag", "error", err)
			goto waitAndContinue
		}
		if isTagActive {
			logging.Info("Cost allocation tag is active")
			return nil
		}

		err = cs.App.AWS.ActivateCostAllocationTag()
		if err != nil {
			logging.Error("Error activating cost allocation tag", "error", err)
			goto waitAndContinue
		}

		//backfill the last month
		err = cs.App.AWS.BackfillCostAllocationTag(time.Now().AddDate(0, -1, 0))
		if err != nil {
			logging.Error("Error backfilling cost allocation tag", "error", err)
			goto waitAndContinue
		}
	waitAndContinue:
		time.Sleep(10 * time.Minute)
	}
}

func (cs *CostWatcher) createBill(unbilledMonth model.YearMonth) error {
	_, err := queries.GetBillByBillingMonth(cs.App.Db, unbilledMonth)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		logging.Error("Error getting bill", "error", err)
		return err
	}

	_, err = queries.GetExploredCost(cs.App.Db, unbilledMonth)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		logging.Error("Error getting explored cost", "error", err)
		return err
	}

	t, err := cs.App.AWS.GetCostForTimeframe(unbilledMonth.Timeframe())
	if err != nil {
		logging.Error("Error getting cost for time period", "error", err)
		return err
	}
	logging.Info("billing", "info", t)

	if len(t) != 1 || t[0] == nil || t[0].BillingMonth != unbilledMonth {
		return fmt.Errorf("cost watcher: error getting cost for time period. Too many results: %v", t)
	}

	err = queries.AddExploredCost(cs.App.Db, unbilledMonth)
	if err != nil {
		logging.Error("Error adding explored cost", "error", err)
		return err
	}

	cost := t[0]
	if cost.Total.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
		logging.Info("No cost for this month", "info", cost)
		return nil
	}

	timeframe := model.Timeframe{
		Start: unbilledMonth.StartTime(),
		End:   unbilledMonth.EndTime(),
	}

	tries := 3
	for t := range tries {
		var cpuTime map[int64]time.Duration
		var playtime map[int64]time.Duration
		activePlayers := map[int64]bool{}
		var serverUptime time.Duration
		var totalCPUTime time.Duration
		var bill *model.Bill
		billItems := make([]*model.BillItem, 0)
		billedPlayers := make([]*model.BilledPlayer, 0)

		players, err := queries.GetPlayers(cs.App.Db)
		if err != nil {
			logging.Error("Error getting players", "error", err)
			goto waitAndContinue
		}

		for p := range players {
			player := players[p]
			if !player.IsAllowed {
				continue
			}
			activePlayers[player.Id] = true
		}
		if len(activePlayers) == 0 {
			logging.Error("Error !! no active players in DB !! Cannot create bill")
			break
		}

		cpuTime, err = queries.CalculateCPUTime(cs.App.Db, timeframe)
		if err != nil {
			logging.Error("Error getting cpu time per player", "error", err)
			goto waitAndContinue
		}
		if cpuTime == nil {
			logging.Info("No cpu time for this month")
			cpuTime = map[int64]time.Duration{}
		}

		playtime, err = queries.CalculatePlaytime(cs.App.Db, timeframe)
		if err != nil {
			logging.Error("Error getting playtime per player", "error", err)
			goto waitAndContinue
		}
		if playtime == nil {
			logging.Info("No playtime for this month")
			playtime = map[int64]time.Duration{}
		}

		serverUptime, err = queries.CalculateServerUptime(cs.App.Db, timeframe)
		if err != nil {
			logging.Error("Error getting server time", "error", err)
			goto waitAndContinue
		}

		for playerId, duration := range cpuTime {
			if ok := activePlayers[playerId]; ok && duration > 0 {
				totalCPUTime += duration
			}
		}

		for p := range players {
			player := players[p]
			if !player.IsAllowed {
				continue
			}

			playerCPUTime, ok := cpuTime[player.Id]
			if !ok {
				playerCPUTime = 0
			}
			playerPlaytime, ok := playtime[player.Id]
			if !ok {
				playerPlaytime = 0
			}

			playerFixedCost := cost.Fixed.Div(decimal.NewFromInt(int64(len(activePlayers))))
			playerVariableCost := cost.Variable.Div(decimal.NewFromInt(int64(len(activePlayers))))
			if totalCPUTime > 0 {
				playerCPUShare := decimal.NewFromFloat(playerCPUTime.Seconds()).Div(decimal.NewFromFloat(totalCPUTime.Seconds()))
				playerVariableCost = cost.Variable.Mul(playerCPUShare)
			}

			totalCost := playerFixedCost.Add(playerVariableCost)
			billledPlayer := &model.BilledPlayer{
				PlayerId:       player.Id,
				TotalAmount:    totalCost,
				FixedAmount:    playerFixedCost,
				VariableAmount: playerVariableCost,
				Playtime:       playerPlaytime,
				CPUTime:        playerCPUTime,
			}

			logging.Debug("Player cost", "player", player.Aliases[0], "totalCost", totalCost, "fixedCost", playerFixedCost, "variableCost", playerVariableCost)
			billedPlayers = append(billedPlayers, billledPlayer)
		}

		for i := range cost.Items {
			costItem := cost.Items[i]
			billItem := &model.BillItem{
				Label:   costItem.Label,
				IsFixed: costItem.IsFixed,
				Amount:  costItem.Amount,
			}
			billItems = append(billItems, billItem)
		}

		bill = &model.Bill{
			BillingMonth:   unbilledMonth,
			TotalAmount:    cost.Total,
			FixedAmount:    cost.Fixed,
			VariableAmount: cost.Variable,
			ServerUptime:   serverUptime,
			Currency:       cost.Currency,
		}

		err = queries.CreateBill(cs.App.Db, bill, billItems, billedPlayers)
		if err != nil {
			logging.Error("Error creating bill", "error", err)
			goto waitAndContinue
		}

		logging.Info("Bill created", "billingMonth", unbilledMonth.String(), "totalCost", cost.Total, "fixedCost", cost.Fixed, "variableCost", cost.Variable)
		break
	waitAndContinue:
		logging.Warn("Error creating bill, retrying", "triesLeft", tries-t-1)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func (cs *CostWatcher) createMissingBills() error {
	var unbilledMonth model.YearMonth
	latestBill, err := queries.GetLatestBill(cs.App.Db)
	if err == sql.ErrNoRows {
		unbilledMonth = model.NewYearMonthFromTime(time.Now()).AddMonths(-1)
	} else if err != nil {
		logging.Error("Error getting latest bill", "error", err)
		return err
	} else {
		unbilledMonth = latestBill.BillingMonth.AddMonths(1)
	}

	currentMonth := model.NewYearMonthFromTime(time.Now())
	for {
		if unbilledMonth.EndTime().After(currentMonth.StartTime()) {
			break
		}

		err = cs.createBill(unbilledMonth)
		if err != nil {
			logging.Error("Error creating bill", "error", err)
			return err
		}
		unbilledMonth = unbilledMonth.AddMonths(1)
	}

	return nil
}

func (cs *CostWatcher) CurrentCostEstimate() (model.CostEstimate, error) {
	return model.CostEstimate{}, nil
}

func (cs *CostWatcher) updateCostEstimate() error {
	//TODO: create new cost estimate if there is data to use
	//get last months fixed and variable cost

	//get last month servertime and playtime
	/* timeframe := model.Timeframe{
		Start: model.NewYearMonthFromTime(time.Now()).StartTime(),
		End:   time.Now(),
	} */

	//get last variable cost
	//get last months fixed cost

	return nil
}

func (cs *CostWatcher) StartWatching() {
	cs.mutex.Lock()
	if cs._isWatching {
		cs.mutex.Unlock()
		logging.Error("CostWatcher is already running")
		return
	}
	logging.Info("Starting CostWatcher")
	cs._isWatching = true
	cs.mutex.Unlock()

	cs.setupCostAllocationTag()

	for {
		var err error
		lastCostUpdate := cs.LastCostUpdate()
		lastCostUpdate = lastCostUpdate.Truncate(time.Minute)

		isHourlyBillingUpdateNeeded := time.Since(lastCostUpdate) >= 1*time.Hour

		//TODO: wait if the logWatcher is still catching up
		isLogwatcherCatchingUp := false

		err = cs.updateCostEstimate()
		if err != nil {
			logging.Error("Error updating cost estimate", "error", err)
		}
		if !isHourlyBillingUpdateNeeded || isLogwatcherCatchingUp {
			goto waitAndContinue
		}
		logging.Debug("CostWatcher: creating missing bills")

		cs.mutex.Lock()
		cs._lastCostUpdate = time.Now()
		cs.mutex.Unlock()

		err = cs.createMissingBills()
		if err != nil {
			logging.Error("Error creating bill for last month", "error", err)
			goto waitAndContinue
		}

		err = cs.App.IS.SyncBills()
		if err != nil {
			logging.Error("Error syncing bills with invoice service", "error", err)
			goto waitAndContinue
		}

	waitAndContinue:
		time.Sleep(1 * time.Minute)
	}
}
