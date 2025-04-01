package serverwatch

import (
	"mc_analyser/db/queries"
	"mc_analyser/flyio"
	"mc_analyser/logging"
	"mc_analyser/model"
	"time"
)

func ServerWatcher(App *model.App) {
	machineAPI := flyio.NewAPI(App.Config.FLY_API_BASE_URL, App.Config.FLY_API_TOKEN, App.Config.FLY_APP_NAME)
	lastEvent, err := queries.GetLastServerEvent(App.Db)
	if err != nil {
		logging.Error("Error getting last server event: ", "error", err)
	}

	lastState := false
	if lastEvent != nil {
		lastState = lastEvent.IsOnline
	}
	App.SetServerOnline(lastState)

	for {
		time.Sleep(5 * time.Second)

		isOnline, err := machineAPI.RestAPI.IsServerOnline()
		if err != nil {
			logging.Error("Error checking server status: ", "error", err)
			continue
		}
		if isOnline == lastState {
			continue
		}

		logging.Info("Server status changed: ", "isOnline", isOnline)
		lastState = isOnline
		App.SetServerOnline(isOnline)

		queries.AddServerEvent(App.Db, time.Now(), isOnline)
	}
}
