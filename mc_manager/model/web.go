package model

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
)

type AuthClaims struct {
	SessionId string
	PlayerId  int64
	jwt.RegisteredClaims
}

// Valid implements the jwt.Claims interface
func (c *AuthClaims) Valid() error {
	if c.ExpiresAt != nil && c.ExpiresAt.Time.Before(time.Now()) {
		return errors.New("token has expired")
	}
	return nil
}

type ServerStatus int

const (
	SERVER_STATUS_ONLINE ServerStatus = iota
	SERVER_STATUS_STARTING
	SERVER_STATUS_OFFLINE
)

type BillResponse struct {
	Month          YearMonth       `json:"billing_month"`
	TotalAmount    decimal.Decimal `json:"total_cost"`
	FixedAmount    decimal.Decimal `json:"fixed_cost"`
	VariableAmount decimal.Decimal `json:"variable_cost"`
	ServerUptime   time.Duration   `json:"server_uptime"`
}

type BillPlayerResponse struct {
	PlayerId       int64           `json:"player_id"`
	Aliases        []string        `json:"aliases"`
	TotalAmount    decimal.Decimal `json:"total_cost"`
	FixedAmount    decimal.Decimal `json:"fixed_cost"`
	VariableAmount decimal.Decimal `json:"variable_cost"`
	Playtime       time.Duration   `json:"playtime"`
	CPUTime        time.Duration   `json:"cpu_time"`
}

type TotalPlayerTimeResponse struct {
	PlayerId int64         `json:"player_id"`
	Aliases  []string      `json:"aliases"`
	Playtime time.Duration `json:"playtime"`
	CPUTime  time.Duration `json:"cpu_time"`
}

type PlayerTimelineResponse struct {
	PlayerId int64          `json:"player_id"`
	Aliases  []string       `json:"aliases"`
	Events   []*PlayerEvent `json:"events"`
}

type EventListResponse struct {
	Players map[int64]*Player
	Events  []*PlayerEvent
}

type SelectorOption struct {
	Label    string
	Url      string
	Target   string
	Swap     string
	Style    string
	Disabled bool
}

var MonthsLabels = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
