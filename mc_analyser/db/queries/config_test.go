package queries_test

import (
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	// Test getting the config
	config, err := queries.GetConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, config.MCServerPort, 25565)
	assert.Nil(t, config.OfflineImage)
	assert.Nil(t, config.OfflineMOTD)

	// Test updating the config
	image := []byte("test_image")
	err = queries.UpdateOfflineImage(DB, &image)
	assert.NoError(t, err)
	config, err = queries.GetConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, config.OfflineImage, &image)

	motd := "test_motd"
	err = queries.UpdateOfflineMOTD(DB, &motd)
	assert.NoError(t, err)
	config, err = queries.GetConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, config.OfflineMOTD, &motd)

	err = queries.UpdateIPCheck(DB, false)
	assert.NoError(t, err)
	config, err = queries.GetConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, false, config.IPCheck)

	err = queries.UpdateIPCheck(DB, true)
	assert.NoError(t, err)
	config, err = queries.GetConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, true, config.IPCheck)

	// Test if length check works
	longMotd := strings.Repeat("a", 1025)
	err = queries.UpdateOfflineMOTD(DB, &longMotd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
}
