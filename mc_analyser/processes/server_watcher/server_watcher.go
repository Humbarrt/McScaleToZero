package server_watcher

import (
	"fmt"
	mc "mc_analyser/api/mc_networking"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type serverWatcher struct {
	App     *model.App
	retries int

	//protected
	mutex       sync.Mutex
	_try        int
	_isWatching bool
	_isOnline   bool
	_isReady    bool
	_ipv4       *string
}

func NewServerWatcher(app *model.App) *serverWatcher {
	return &serverWatcher{
		App:     app,
		retries: 5,
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
	sw._try = 0
	sw.mutex.Unlock()
	if isWatching {
		logging.Error("ServerWatcher is already running")
		return
	}

	sw.mutex.Lock()
	sw._isWatching = true
	sw.mutex.Unlock()

	logging.Info("Starting ServerWatcher")
	for {
		sw.mutex.Lock()
		wasOnline := sw._isOnline
		wasReady := sw._isReady
		retries := sw._try

		if retries > sw.retries {
			sw._isWatching = false
			sw.mutex.Unlock()
			break
		}
		sw.mutex.Unlock()

		isOnline := false
		isStateStable := true
		state, err := sw.App.AWS.GetInstanceState()
		if err != nil {
			logging.Error("Error checking server state: ", "error", err)
			goto waitAndContinue
		}
		isOnline = state == ec2Types.InstanceStateNameRunning
		isStateStable = state == ec2Types.InstanceStateNameRunning || state == ec2Types.InstanceStateNameStopped || state == ec2Types.InstanceStateNameTerminated
		logging.Debug("Server state: ", "isOnline", isOnline, "wasOnline", wasOnline, "wasReady", wasReady, "state", state, "isStateStable", isStateStable)

		if !wasOnline && isOnline {
			logging.Info("Server started")
			ipv4, err := sw.App.AWS.GetInstanceIP()
			if err != nil {
				logging.Error("Error getting server IP: ", "error", err)
				goto waitAndContinue
			}

			sw.mutex.Lock()
			sw._try = 0
			sw._ipv4 = &ipv4
			sw._isOnline = true
			sw.mutex.Unlock()
			logging.Info("Server is set to online")

			go UpdateWhitelistAndAdmins(sw.App)
			go BackupWorld(sw.App)
		}

		if isOnline && !wasReady {
			logging.Info("Server is online but not ready")

			ipv4, err := sw.App.AWS.GetInstanceIP()
			if err != nil {
				logging.Error("Server is online but has no IP: ", "error", err)
				logging.Warn("Starting IP recovery")

				ipv4, err = sw.App.AWS.GetInstanceIP()
				if err != nil {
					logging.Error("IP recovery failed: ", "error", err)
					goto waitAndContinue
				}
				sw.mutex.Lock()
				sw._ipv4 = &ipv4
				sw.mutex.Unlock()

				logging.Info("IP recovery successful")
			}

			canTransfer, err := mc.IsServerTransferReady(ipv4, sw.App.Config.GetMCPort())
			if err != nil || !canTransfer {
				logging.Error("Error reading server status: ", "error", err, "canTransfer", canTransfer)
				goto waitAndContinue
			}

			sw.mutex.Lock()
			sw._isReady = true
			sw.mutex.Unlock()
			logging.Info("Server is ready")
		}

		if wasOnline && !isOnline {
			logging.Info("Server is stopping")

			sw.mutex.Lock()
			sw._isOnline = false
			sw._isReady = false
			sw._ipv4 = nil
			sw.mutex.Unlock()
		}

		if !isOnline && isStateStable {
			logging.Info("Server is offline", "retries left", sw.retries-retries)
			sw.mutex.Lock()
			sw._try = retries + 1
			sw.mutex.Unlock()
		}

	waitAndContinue:
		time.Sleep(5 * time.Second)
	}

	logging.Info("ServerWatcher stopped")
}
