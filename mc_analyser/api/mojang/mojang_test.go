package mojang_test

import (
	"mc_analyser/api/mojang"
	"mc_analyser/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlayerUUID(t *testing.T) {
	playerName := "Notch"
	expectedUUID, err := model.UUIDFromString("069a79f4-44e9-4726-a5be-fca90e38aaf5")
	assert.NoError(t, err)

	uuid, err := mojang.GetPlayerUUID(playerName)
	assert.NoError(t, err)
	assert.Equal(t, expectedUUID.Bytes(), uuid.Bytes())
}
