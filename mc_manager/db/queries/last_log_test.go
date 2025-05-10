package queries_test

import (
	"database/sql"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/model"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLastLog(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	// Get empty last log
	lastLog, err := queries.GetLastLog(DB)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, lastLog)

	// Create a new log
	log := &model.Log{
		Timestamp: time.Now(),
		Message:   "Test log message",
	}

	// Set the last log
	err = queries.SetLastLog(DB, log)
	assert.NoError(t, err)

	// Get the last log
	lastLog, err = queries.GetLastLog(DB)
	assert.NoError(t, err)

	// Check if the last log is the same as the one we set
	assert.Equal(t, log.Timestamp.Unix(), lastLog.Timestamp.Unix())
	assert.Equal(t, log.Message, lastLog.Message)
}
