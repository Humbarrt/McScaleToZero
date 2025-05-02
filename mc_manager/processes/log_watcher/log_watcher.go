package log_watcher

import (
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"time"
)

func LogWatcher(App *model.App) {
	lastState := App.SW.IsServerOnline()
	for {
		time.Sleep(1 * time.Second)
		isOnline := App.SW.IsServerOnline()

		hasStarted := isOnline && lastState != isOnline
		hasStopped := !isOnline && lastState != isOnline
		lastState = isOnline

		if hasStarted {
			// Server has started
			// TODO: download whitelist.json
			// TODO: load 100 lines with prometheus
			// TODO: if last lines not in prometheus, load files
		} else if hasStopped {
			events, err := queries.GetLastestPlayersEvents(App.Db)
			if err != nil {
				logging.Error("Error getting last player events", "error", err)
				continue
			}

			for _, event := range events {
				if event.IsOnline {
					continue
				}

				err = queries.AddPlayerEvent(App.Db, time.Now().UTC(), event.PlayerId, false)
				if err != nil {
					logging.Error("Error adding player event", "error", err)
					continue
				}
			}

			logging.Info("Server stopped, set all players to offline")
		} else if isOnline {
			// Server is online
			// TODO: load 100 lines with prometheus
			time.Sleep(4 * time.Second)
		}
	}
}
