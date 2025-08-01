package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ServerWatcher interface {
	IsServerOnline() bool
	IsServerReady() bool
	StartWatchingServer()
	IPv4() (string, error)
	LastSeen() time.Time
}

type LogWatcher interface {
	StartLogWatcher()
	IsCatchingUp() bool
}

type CostWatcher interface {
	StartWatching()
	CurrentCostEstimate() (CostEstimate, error)
}

type Invoice struct {
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	BillId    int       `json:"bill_id"`
	TotalCost int       `json:"total_cost"`
}

type CostEstimate struct {
	TotalCost    decimal.Decimal `json:"total_cost"`
	FixedCost    decimal.Decimal `json:"fixed_cost"`
	VariableCost decimal.Decimal `json:"variable_cost"`
	Currency     string          `json:"currency"`
	ServerUptime time.Duration   `json:"server_uptime"`
	Players      []PlayerCostEstimate
}

type PlayerCostEstimate struct {
	PlayerId     int64           `json:"player_id"`
	TotalCost    decimal.Decimal `json:"cost"`
	FixedCost    decimal.Decimal `json:"fixed_cost"`
	VariableCost decimal.Decimal `json:"variable_cost"`
	Playtime     time.Duration   `json:"playtime"`
	CPUTime      time.Duration   `json:"cpu_time"`
}
