package queries

import (
	"database/sql"
	"math/rand"
	"mc_manager/db"
	"mc_manager/model"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetExploredCost(t *testing.T) {
	filePath := path.Join(t.TempDir(), "main.db")
	DB, err := db.InitializeDatabase(filePath)
	assert.NoError(t, err)

	//add explored cost
	startDate := model.NewYearMonthFromTime(time.Now()).StartTime()
	exploredMonths := map[model.YearMonth]bool{}
	for i := range 24 {
		if rand.Intn(100) < 50 {
			continue
		}

		ym := model.NewYearMonthFromTime(startDate.AddDate(0, i, 0))
		err = AddExploredCost(DB, ym)
		assert.NoError(t, err)
		exploredMonths[ym] = true
	}

	//get explored cost
	for i := range 24 {
		ym := model.NewYearMonthFromTime(startDate.AddDate(0, i, 0))
		cost, err := GetExploredCost(DB, ym)
		if exploredMonths[ym] {
			assert.NoError(t, err)
			assert.Equal(t, ym, cost.ExploredMonth)
		} else {
			assert.Error(t, sql.ErrNoRows, err)
		}
	}
}
