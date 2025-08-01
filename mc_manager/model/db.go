package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Player struct {
	Id                   int64    `json:"id"`
	McUUID               UUID     `json:"mc_uuid"`
	IsAllowed            bool     `json:"is_active"`
	IsAdmin              bool     `json:"is_admin"`
	LastIPV4             *string  `json:"last_ip"`
	Aliases              []string `json:"aliases"`
	InvoiceServiceUserId *string  `json:"invoice_service_user_id"`
}

type ExploredCost struct {
	Id            int64
	CreatedAt     time.Time
	ExploredMonth YearMonth
}

type PlayerEvent struct {
	Id        int64     `json:"id"`
	PlayerId  int64     `json:"player_id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type ServerEvent struct {
	Id        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type Playtime struct {
	PlayerId int64         `json:"player_id"`
	Time     time.Duration `json:"playtime"`
}

type CPUTime struct {
	PlayerId int64         `json:"player_id"`
	Time     time.Duration `json:"playtime"`
}

type Bill struct {
	Id                      int64           `json:"id"`
	CreatedAt               time.Time       `json:"created_at"`
	BillingMonth            YearMonth       `json:"billing_month"`
	TotalAmount             decimal.Decimal `json:"total_cost"`
	FixedAmount             decimal.Decimal `json:"fixed_cost"`
	VariableAmount          decimal.Decimal `json:"variable_cost"`
	ServerUptime            time.Duration   `json:"server_uptime"`
	Currency                string          `json:"currency"`
	InvoiceServiceInvoiceId *string         `json:"invoice_service_invoice_id"`
}

type BillItem struct {
	Id      int64           `json:"id"`
	BillId  int64           `json:"bill_id"`
	Label   string          `json:"label"`
	IsFixed bool            `json:"is_fixed"`
	Amount  decimal.Decimal `json:"cost"`
}

type BilledPlayer struct {
	Id             int64           `json:"id"`
	BillId         int64           `json:"bill_id"`
	PlayerId       int64           `json:"player_id"`
	TotalAmount    decimal.Decimal `json:"cost"`
	FixedAmount    decimal.Decimal `json:"fixed_cost"`
	VariableAmount decimal.Decimal `json:"variable_cost"`
	Playtime       time.Duration   `json:"playtime"`
	CPUTime        time.Duration   `json:"cpu_time"`
}

type FullBill struct {
	Bill          *Bill           `json:"bill"`
	BillItems     []*BillItem     `json:"bill_items"`
	BilledPlayers []*BilledPlayer `json:"billed_players"`
}

type BilledPlayerTime struct {
	PlayerId int64         `json:"player_id"`
	Playtime time.Duration `json:"playtime"`
	CPUTime  time.Duration `json:"cpu_time"`
}
