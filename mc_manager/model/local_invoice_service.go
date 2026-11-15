package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// hard-coded currencies for now
var LocalInvoiceServiceSupportedCurrencies = map[string]string{
	"AUD": "$",
	"BGN": "лв",
	"BRL": "R$",
	"CAD": "$",
	"CHF": "CHF",
	"CNY": "¥",
	"CZK": "Kč",
	"DKK": "kr",
	"EUR": "€ ",
	"GBP": "£",
	"HKD": "$",
	"HRK": "kn",
	"HUF": "Ft",
	"IDR": "Rp",
	"ILS": "₪",
	"INR": "₹",
	"ISK": "kr",
	"JPY": "¥",
	"KRW": "₩",
	"MXN": "$",
	"MYR": "RM",
	"NOK": "kr",
	"NZD": "$",
	"PHP": "₱",
	"PLN": "zł",
	"RON": "lei",
	"RUB": "₽",
	"SEK": "kr",
	"SGD": "$",
	"THB": "฿",
	"TRY": "₺",
	"USD": "$",
	"ZAR": "R",
}

type LocalInvoiceServicePlayerCache struct {
	TotalSpent      decimal.Decimal
	TotalPaid       decimal.Decimal
	OutstandingDebt decimal.Decimal
	PaymentCount    int
}

type LocalInvoiceServicePlayerCacheResponse struct {
	Player         *Player
	Cache          *LocalInvoiceServicePlayerCache
	IsPayingPlayer bool
}

type LocalInvoiceServiceCacheResponse struct {
	Players         []*LocalInvoiceServicePlayerCacheResponse
	TotalSpent      decimal.Decimal
	TotalPaid       decimal.Decimal
	OutstandingDebt decimal.Decimal
	InvoiceCount    int
	PaymentCount    int
}

type LocalInvoiceServicePlayerHistoryResponse struct {
	SortingDate time.Time
	Invoice     *ExtendedLocalInvoiceServiceInvoiceItem
	Payment     *LocalInvoiceServicePayment
}
