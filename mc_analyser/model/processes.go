package model

type ServerWatcher interface {
	IsServerOnline() bool
	IsServerReady() bool
	StartWatchingServer()
	IPv4() (string, error)
}
