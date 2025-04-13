package cost_watcher

import (
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"

	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// request splitwise, for all users in the group and all bills.
// if the bills of db are not in splitwise, create them. --> usefull if the manager failed to create previous invoices.

// on start loop check if the aws allocation tag exists and if it is activated. If not, wait for it (sleep 10 minutes), and then activate it and Backfill it for the current + last month.
// after allocation tag is confirmed, go on.
// if log_watcher is still catching up, wait for it. --> sleep 1 minute.

// enter cost explorer loop
// check the cost explorer, make data accessible to the rest of the app. Split in fixed and variable costs.
// if there is no bill for the last month in db, create it.
// is the server online, if yes sleep 1 minute, if no, sleep 10 minutes.

// create bill: save all cost points in db, get how much cpu time each player was using.
//      if there was no player usage, split the bill equally.
//      if there was player usage, split the bills variable costs by player usage and fixed costs equally.
//      save the bill in db.
//      create a splitwise entry.

// the db should look like:
// bills:
// - id int
// - created_at timestamp
// - billing_month int
// - total_cost int
// - fixed_cost int
// - variable_cost int
// - currency string
// - splitwise_invoice_id string

// bill_items:
// - id
// - bill_id
// - label
// - is_fixed bool
// - cost int

// billed_players:
// - id
// - bill_id
// - player_id
// - cost int
// - playtime int
// - cpu_time int

type costWatcher struct {
	App *model.App

	//protected
	mutex           sync.Mutex
	_isWatching     bool
	_lastCostUpdate time.Time
}

func NewCostWatcher(app *model.App) *costWatcher {
	return &costWatcher{
		App: app,
	}
}

func (cs *costWatcher) LastCostUpdate() time.Time {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	return cs._lastCostUpdate
}

func (cs *costWatcher) setupCostAllocationTag() error {
	for {
		tag, err := cs.App.AWS.GetCostAllocationTag("Project")
		if err != nil || tag == nil {
			logging.Error("Error getting cost allocation tag", "error", err, "tag", tag)
			goto waitAndContinue
		}

		if tag.Status == ceTypes.CostAllocationTagStatusActive {
			logging.Info("Cost allocation tag is active", "tag", tag)
			return nil
		}

		err = cs.App.AWS.ActivateCostAllocationTag("Project")
		if err != nil {
			logging.Error("Error activating cost allocation tag", "error", err)
			goto waitAndContinue
		}

		logging.Info("Cost allocation tag activated", "tag", tag)
		// Backfill from the last month to now
		err = cs.App.AWS.BackfillCostAllocationTag("Project", time.Now().AddDate(0, -1, 0))
		if err != nil {
			logging.Error("Error backfilling cost allocation tag", "error", err)
			goto waitAndContinue
		}

		logging.Info("Cost allocation tag backfilled", "tag", tag)
		return nil

	waitAndContinue:
		time.Sleep(10 * time.Minute)
	}
}

func (cs *costWatcher) StartWatching() {
	cs.mutex.Lock()
	if cs._isWatching {
		cs.mutex.Unlock()
		logging.Error("CostWatcher is already running")
		return
	}
	logging.Info("Starting CostWatcher")
	cs._isWatching = true
	cs.mutex.Unlock()

	//TODO: init splitwise api
	//TODO: ensure all bills have splitwise invoices

	cs.setupCostAllocationTag()

	for {
		lastCostUpdate := cs.LastCostUpdate()
		lastCostUpdate = lastCostUpdate.Truncate(time.Minute)

		isOnline := cs.App.SW.IsServerOnline()
		wasRecentlyOnline := time.Since(cs.App.SW.LastSeen()) < 10*time.Minute
		isHourlyCostUpdateNeeded := time.Since(lastCostUpdate) >= 1*time.Hour

		if !isOnline && !wasRecentlyOnline && !isHourlyCostUpdateNeeded {
			goto waitAndContinue
		}

		//TODO: check if the logWatcher is still catching up
		cs.mutex.Lock()
		cs._lastCostUpdate = time.Now()
		cs.mutex.Unlock()
		//TODO: get the cost from cost explorere
		//TODO: publish current costs
		//TODO: if there is no bill for the last month in db, create it.

		logging.Info("CostWatcher: Updating cost data")

	waitAndContinue:
		time.Sleep(1 * time.Minute)
	}
}
