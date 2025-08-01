package model

import (
	"fmt"
	"time"
)

type YearMonth int

// YearMonth represents a year and month in the format YYYYMM.
// String() returns the year and month in the format "YYYY-MM".
func (ym YearMonth) String() string {
	year := ym / 100
	month := ym % 100
	return fmt.Sprintf("%04d-%02d", year, month)
}

func (ym YearMonth) Year() int {
	return int(ym / 100)
}
func (ym YearMonth) Month() int {
	return int(ym % 100)
}
func (ym YearMonth) StartTime() time.Time {
	return time.Date(ym.Year(), time.Month(ym.Month()), 1, 0, 0, 0, 0, time.UTC)
}
func (ym YearMonth) EndTime() time.Time {
	return ym.StartTime().AddDate(0, 1, 0).Add(-time.Nanosecond)
}
func (ym YearMonth) Timeframe() Timeframe {
	return Timeframe{
		Start: ym.StartTime(),
		End:   ym.EndTime(),
	}
}
func (ym YearMonth) AddMonths(months int) YearMonth {
	return NewYearMonthFromTime(ym.StartTime().AddDate(0, months, 0))
}

func NewYearMonthFromInt(year, month int) (YearMonth, error) {
	if month < 1 || month > 12 {
		return 0, fmt.Errorf("month must be between 1 and 12")
	}
	return YearMonth(year*100 + month), nil
}

func NewYearMonthFromString(s string) (YearMonth, error) {
	var year, month int
	_, err := fmt.Sscanf(s, "%d-%d", &year, &month)
	if err != nil {
		return 0, err
	}

	return NewYearMonthFromInt(year, month)
}

func NewYearMonthFromTime(t time.Time) YearMonth {
	return YearMonth(t.UTC().Year()*100 + int(t.UTC().Month()))
}
