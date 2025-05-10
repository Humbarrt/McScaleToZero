package queries

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal(t *testing.T) {
	// Test the decimal value
	decimalValue := decimal.NewFromFloat(123.4567)
	_int := decimalToInt(decimalValue)
	assert.Equal(t, int64(1234567), _int)
	decimalValue = intToDecimal(_int)
	assert.Equal(t, "123.4567", decimalValue.String())

	decimalValue = decimal.NewFromFloat(123.45671)
	_int = decimalToInt(decimalValue)
	assert.Equal(t, int64(1234568), _int)
	decimalValue = intToDecimal(_int)
	assert.Equal(t, "123.4568", decimalValue.String())

	decimalValue = decimal.NewFromFloat(123.45)
	_int = decimalToInt(decimalValue)
	assert.Equal(t, int64(1234500), _int)
	decimalValue = intToDecimal(_int)
	assert.Equal(t, "123.45", decimalValue.String())

	decimalValue = decimal.NewFromFloat(0)
	_int = decimalToInt(decimalValue)
	assert.Equal(t, int64(0), _int)
	decimalValue = intToDecimal(_int)
	assert.Equal(t, "0", decimalValue.String())
}
