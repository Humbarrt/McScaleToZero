package queries_test

import (
	"bytes"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/model"
	"path"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetMutableConfig(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	// Test getting the config
	config := getConfigNoError(t, DB)
	assert.Equal(t, 25565, config.MCServerPort)
	assert.Nil(t, config.OnlineImage)
	assert.Equal(t, []byte("\u00A7r\n                    \u00A7a(running)"), config.OnlineMOTD)
	assert.Nil(t, config.OfflineImage)
	assert.Equal(t, []byte("\u00A7r\n                \u00A78(not started yet)"), config.OfflineMOTD)

	// Test updating the config
	onlineImage := "test_online_image"
	err = queries.UpdateOnlineImage(DB, &onlineImage)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.NotNil(t, config.OnlineImage)
	assert.Equal(t, *config.OnlineImage, onlineImage)

	onlineMotd := []byte("test_online_motd")
	err = queries.UpdateOnlineMOTD(DB, onlineMotd)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, config.OnlineMOTD, onlineMotd)

	offlineImage := "test_offline_image"
	err = queries.UpdateOfflineImage(DB, &offlineImage)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, config.OfflineImage, &offlineImage)

	offlineMotd := []byte("test_offline_motd")
	err = queries.UpdateOfflineMOTD(DB, offlineMotd)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, config.OfflineMOTD, offlineMotd)

	//ReplaceMOTD
	err = queries.UpdateReplaceMOTD(DB, true)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, true, config.ReplaceMOTD)

	err = queries.UpdateReplaceMOTD(DB, false)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, false, config.ReplaceMOTD)

	//LastImage
	lastImage := "test_last_image"
	err = queries.UpdateLastImage(DB, &lastImage)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, config.LastImage, &lastImage)

	err = queries.UpdateLastImage(DB, nil)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Nil(t, config.LastImage)

	//LastMOTD
	lastMotd := []byte("test_last_motd")
	err = queries.UpdateLastMOTD(DB, lastMotd)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, config.LastMOTD, lastMotd)

	err = queries.UpdateLastMOTD(DB, []byte(""))
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.True(t, bytes.Equal(config.LastMOTD, []byte("")))

	//ProtocolVersion
	protocolVersion := []byte("test_protocol_version")
	err = queries.UpdateProtocolVersion(DB, &protocolVersion)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.True(t, bytes.Equal(*config.ProtocolVersion, protocolVersion))

	err = queries.UpdateProtocolVersion(DB, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error protocol version is nil")
	config = getConfigNoError(t, DB)
	assert.True(t, bytes.Equal(*config.ProtocolVersion, protocolVersion))

	emptyProtocolVersion := []byte("")
	err = queries.UpdateProtocolVersion(DB, &emptyProtocolVersion)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
	config = getConfigNoError(t, DB)
	assert.True(t, bytes.Equal(*config.ProtocolVersion, protocolVersion))

	//MaxPlayers
	err = queries.UpdateMaxPlayers(DB, 20)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, 20, config.MaxPlayers)
	err = queries.UpdateMaxPlayers(DB, 100)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, 100, config.MaxPlayers)
	err = queries.UpdateMaxPlayers(DB, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
	config = getConfigNoError(t, DB)
	assert.Equal(t, 100, config.MaxPlayers)

	// Test if length check works
	longMotd := strings.Repeat("a", 1025)
	err = queries.UpdateOfflineMOTD(DB, []byte(longMotd))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")

	// Test if update backups to keep works
	err = queries.UpdateBackupsToKeep(DB, 5)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, 5, config.BackupsToKeep)

	err = queries.UpdateBackupsToKeep(DB, 15)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, 15, config.BackupsToKeep)

	err = queries.UpdateBackupsToKeep(DB, 0)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.Equal(t, 0, config.BackupsToKeep)

	err = queries.UpdateBackupsToKeep(DB, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
	config = getConfigNoError(t, DB)
	assert.Equal(t, 0, config.BackupsToKeep)

	err = queries.UpdateBackupsToKeep(DB, 16)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
	config = getConfigNoError(t, DB)
	assert.Equal(t, 0, config.BackupsToKeep)

	// Test if update tax rate works
	taxDec := decimal.NewFromFloat(0.2)
	err = queries.UpdateTaxRate(DB, taxDec)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.True(t, taxDec.Equal(config.TaxRate))

	taxDec = decimal.NewFromFloat(0.5)
	err = queries.UpdateTaxRate(DB, taxDec)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.True(t, taxDec.Equal(config.TaxRate))

	taxDec = decimal.NewFromFloat(0.0)
	err = queries.UpdateTaxRate(DB, taxDec)
	assert.NoError(t, err)
	config = getConfigNoError(t, DB)
	assert.True(t, taxDec.Equal(config.TaxRate))

	taxDec = decimal.NewFromFloat(-0.1)
	err = queries.UpdateTaxRate(DB, taxDec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CHECK constraint failed")
	config = getConfigNoError(t, DB)
	assert.True(t, decimal.NewFromFloat(0.0).Equal(config.TaxRate))
}

func getConfigNoError(t *testing.T, DB *db.DB) *model.MutableConfig {
	config, err := queries.GetMutableConfig(DB)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	return config
}
