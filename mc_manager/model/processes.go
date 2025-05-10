package model

import "time"

type ServerWatcher interface {
	IsServerOnline() bool
	IsServerReady() bool
	StartWatchingServer()
	IPv4() (string, error)
	LastSeen() time.Time
}

type CostWatcher interface {
	StartWatching()
	CurrentCostEstimate() (*CostEstimate, error)
}

type LogWatcher interface {
	StartLogWatcher()
}
