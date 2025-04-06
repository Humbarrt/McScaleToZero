package model_test

import (
	"mc_analyser/model"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	// Test UUID creation from long string
	uuidStr := "c99efdca-519c-43ea-83fb-ebc7042c8e3f"
	uuid, err := model.NewUUIDFromString(uuidStr)
	assert.NoError(t, err)
	assert.Equal(t, uuid.LongString(), uuidStr)
	assert.Equal(t, uuid.ShortString(), strings.ReplaceAll(uuidStr, "-", ""))
	assert.Equal(t, uuid.Bytes(), []byte{0xc9, 0x9e, 0xfd, 0xca, 0x51, 0x9c, 0x43, 0xea, 0x83, 0xfb, 0xeb, 0xc7, 0x04, 0x2c, 0x8e, 0x3f})

	// Test UUID creation from bytes
	uuid2, err := model.NewUUIDFromBytes(uuid.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, uuid2.Bytes(), uuid.Bytes())

	// Test UUID creation from short string
	uuidStrShort := strings.ReplaceAll(uuidStr, "-", "")
	uuid3, err := model.NewUUIDFromString(uuidStrShort)
	assert.NoError(t, err)
	assert.Equal(t, uuid3.Bytes(), uuid.Bytes())

	// Test invalid UUID string
	invalidUUIDStr := "c99efdca-519c-43ea-83fb-ebc7042c8e3"
	_, err = model.NewUUIDFromString(invalidUUIDStr)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "UUID must be 32 characters long")

	// Test invalid UUID bytes
	invalidUUIDBytes := []byte{0xc9, 0x9e, 0xfd, 0xca, 0x51, 0x9c, 0x43, 0xea, 0x83, 0xfb, 0xeb, 0xc7, 0x04, 0x2c}
	_, err = model.NewUUIDFromBytes(invalidUUIDBytes)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "UUID must be 16 bytes long")

	// Test that two unique UUIDs are not equal
	uuid4, err := model.NewUUIDFromString("c99efdca-519c-43ea-83fb-ebc7042c8e3g")
	assert.NoError(t, err)
	assert.NotEqual(t, uuid, uuid4)
}
