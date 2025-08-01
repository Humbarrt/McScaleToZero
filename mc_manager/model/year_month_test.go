package model_test

import (
	"mc_manager/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestYearMonth(t *testing.T) {
	// Test NewYearMonthFromTime
	ym := model.NewYearMonthFromTime(time.Date(2023, 3, 12, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, 2023, ym.Year())
	assert.Equal(t, 3, ym.Month())

	// Test NewYearMonth
	ym, err := model.NewYearMonthFromInt(2023, 3)
	assert.NoError(t, err)
	assert.Equal(t, 2023, ym.Year())
	assert.Equal(t, 3, ym.Month())
	assert.Equal(t, 202303, int(ym))

	// Test NewYearMonthFromString
	ym, err = model.NewYearMonthFromString("2023-10")
	assert.NoError(t, err)
	assert.Equal(t, 2023, ym.Year())
	assert.Equal(t, 10, ym.Month())

	// Test String method
	assert.Equal(t, "2023-10", ym.String())

	// Test StartTime and EndTime methods
	startTime := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	assert.Equal(t, startTime, ym.StartTime())
	assert.Equal(t, endTime, ym.EndTime())

	// Test Timeframe method
	tf := ym.Timeframe()
	assert.Equal(t, startTime, tf.Start)
	assert.Equal(t, endTime, tf.End)

	// Test AddMonths method
	ym2 := ym.AddMonths(1)
	assert.Equal(t, 2023, ym2.Year())
	assert.Equal(t, 11, ym2.Month())
	assert.Equal(t, 202311, int(ym2))
	ym3 := ym.AddMonths(-1)
	assert.Equal(t, 2023, ym3.Year())
	assert.Equal(t, 9, ym3.Month())
	assert.Equal(t, 202309, int(ym3))

	ym4 := ym.AddMonths(3)
	assert.Equal(t, 2024, ym4.Year())
	assert.Equal(t, 1, ym4.Month())
	assert.Equal(t, 202401, int(ym4))
}
