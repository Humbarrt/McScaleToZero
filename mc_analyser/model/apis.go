package model

import (
	"time"

	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/shopspring/decimal"
)

type CostItem struct {
	Label   string          `json:"label"`
	Amount  decimal.Decimal `json:"cost"`
	IsFixed bool            `json:"is_fixed"`
}

type MonthsCost struct {
	Timeframe    Timeframe       `json:"timeframe"`
	BillingMonth YearMonth       `json:"billing_month"`
	Total        decimal.Decimal `json:"total"`
	Fixed        decimal.Decimal `json:"fixed"`
	Variable     decimal.Decimal `json:"variable"`
	Currency     string          `json:"currency"`
	Items        []CostItem      `json:"items"`
}

type AWSAPI interface {
	StartInstance() error
	GetInstanceState() (ec2Types.InstanceStateName, error)
	IsInstanceOnline() (bool, error)
	GetInstanceIP() (string, error)
	GetCostAllocationTag() (*ceTypes.CostAllocationTag, error)
	IsCostAllocationTagActive() (bool, error)
	ActivateCostAllocationTag() error
	BackfillCostAllocationTag(backfillFrom time.Time) error
	IsBackfillInProgress() (bool, error)
	GetCostForTimeframe(timeframe Timeframe) ([]*MonthsCost, error)
}

type InvoiceService string

const (
	InvoiceServiceNone      InvoiceService = "none"
	InvoiceServiceSplitwise InvoiceService = "splitwise"
)

type InvoiceServiceStatus int

// This is a bit overkill right now, but good to have incase of oauth implementations...
const (
	StatusInvoiceServiceNotInitialized InvoiceServiceStatus = iota
	StatusInvoiceServicePending
	StatusInvoiceServiceReady
	StatusInvoiceServiceError
)

type InvoiceServiceAPI interface {
	Init() error
	Status() InvoiceServiceStatus

	SyncBills() error
	GetUsers() ([]*InvoiceServiceUser, error)
}
