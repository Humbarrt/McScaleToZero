package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
)

func HumanizeMoney(amount decimal.Decimal) string {
	amountFloat, _ := amount.RoundCeil(2).Float64()
	content := humanize.Commaf(amountFloat)
	dotPos := strings.LastIndex(content, ".")
	if dotPos == -1 {
		content += ".00"
	} else if len(content)-dotPos == 2 {
		content += "0"
	}
	return content
}

func HumanizeDuration(dur time.Duration) string {
	content := ""
	years := int(dur.Hours() / 24 / 365)
	days := int(dur.Hours()/24) % 365
	hours := int(dur.Hours()) % 24
	minutes := int(dur.Minutes()) % 60
	seconds := int(dur.Seconds()) % 60
	if years > 0 {
		content += fmt.Sprintf("%dy ", years)
	}
	if days > 0 {
		content += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		content += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		content += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 {
		content += fmt.Sprintf("%ds ", seconds)
	}
	if content == "" {
		content = "0s"
	}

	return content
}
