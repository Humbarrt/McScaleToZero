package queries

import "github.com/shopspring/decimal"

type scannable interface {
	Scan(dest ...any) error
}

func decimalToInt(dec decimal.Decimal) int64 {
	dec = dec.Round(4)
	dec = dec.Mul(decimal.NewFromInt(10000))
	return dec.IntPart()
}

func intToDecimal(i int64) decimal.Decimal {
	dec := decimal.NewFromInt(i)
	return dec.Div(decimal.NewFromInt(10000))
}
