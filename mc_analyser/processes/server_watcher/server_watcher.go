package server_watcher

import (
	"fmt"
	mc "mc_analyser/api/mcnetworking"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"
)

type serverWatcher struct {
	App     *model.App
	retries int

	//protected
	mutex       sync.Mutex
	_isWatching bool
	_isOnline   bool
	_isReady    bool
	_ipv4       *string
}

func NewServerWatcher(app *model.App) *serverWatcher {
	return &serverWatcher{
		App: app,
	}
}

func (sw *serverWatcher) IsServerOnline() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	return sw._isOnline
}

func (sw *serverWatcher) IsServerReady() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	return sw._isReady
}

func (sw *serverWatcher) IPv4() (string, error) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	if sw._ipv4 == nil {
		return "", fmt.Errorf("IPv4 address is not set")
	}
	return *sw._ipv4, nil
}

func (sw *serverWatcher) StartWatchingServer() {
	logging.Info("Trying to start ServerWatcher")
	sw.mutex.Lock()
	isWatching := sw._isWatching
	sw.mutex.Unlock()
	if isWatching {
		logging.Error("ServerWatcher is already running")
		return
	}

	sw._isWatching = true
	sw.retries = 0
	logging.Info("Starting ServerWatcher")
	for sw.retries <= sw.App.Config.SERVER_WATCHER_RETRIES {
		sw.mutex.Lock()
		wasOnline := sw._isOnline
		wasReady := sw._isReady
		sw.mutex.Unlock()

		isOnline, err := sw.App.AWS.IsInstanceOnline()
		if err != nil {
			logging.Error("Error checking server status: ", "error", err)
			goto waitAndContinue
		}
		logging.Info("Server status: ", "isOnline", isOnline, "wasOnline", wasOnline)

		if !wasOnline && isOnline {
			logging.Info("Server started")
			sw.retries = 0

			ipv4, err := sw.App.AWS.GetInstanceIP()
			if err != nil {
				logging.Error("Error getting server IP: ", "error", err)
				goto waitAndContinue
			}

			sw.mutex.Lock()
			sw._ipv4 = &ipv4
			sw._isOnline = true
			sw.mutex.Unlock()
			logging.Info("Server is set to online")
			logging.Info("Server has IP: ", "IP", ipv4)

			//TODO: get whitelist
			//TODO: make server backup
		}

		if isOnline && !wasReady {
			logging.Info("Server is online but not ready")

			ipv4, err := sw.App.AWS.GetInstanceIP()
			if err != nil {
				logging.Error("Server is online but has no IP: ", "error", err)
				logging.Info("Trying to recover from state")

				ipv4, err = sw.App.AWS.GetInstanceIP()
				if err != nil {
					logging.Error("Error getting server IP: ", "error", err)
					goto waitAndContinue
				}
				sw.mutex.Lock()
				sw._ipv4 = &ipv4
				sw.mutex.Unlock()

				logging.Info("Recovery successful")
				logging.Info("Server has IP: ", "IP", ipv4)
			}

			status, err := mc.ReadServerStatus(ipv4, sw.App.Config.MC_PORT)
			if err != nil || status == "" {
				logging.Error("Error reading server status: ", "error", err, "status", status)
				goto waitAndContinue
			}

			sw.mutex.Lock()
			sw._isReady = true
			sw.mutex.Unlock()
			logging.Info("Server is ready")
		}

		if wasOnline && !isOnline {
			logging.Info("Server stopped")

			sw.mutex.Lock()
			sw._isOnline = false
			sw._isReady = false
			sw._ipv4 = nil
			sw.mutex.Unlock()
		}

		if !isOnline {
			logging.Info("Server is offline", "retries", sw.retries)
			sw.retries += 1
		}

	waitAndContinue:
		time.Sleep(5 * time.Second)
	}

	sw.mutex.Lock()
	sw._isWatching = false
	sw.mutex.Unlock()
	logging.Info("ServerWatcher stopped")
}
