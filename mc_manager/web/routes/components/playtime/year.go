package playtime

import (
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	playtime_view "mc_manager/web/views/components/playtime"
	utils "mc_manager/web/web_utils"
	"time"

	"github.com/gin-gonic/gin"
)

func renderYearPlaytime(ctx *gin.Context, app *model.App, year int) {
	start := year * 100
	end := start + 99
	billedPlayerTime, err := queries.GetBilledPlayerTimeForTimespan(app.Db, start, end)
	if err != nil {
		logging.Error("Error getting billed player time", "error", err)
		ctx.Status(500)
		return
	}

	isLive := false
	var currentPlayerEstimate *[]model.PlayerCostEstimate
	if year == time.Now().Year() {
		_currentEstimate, err := app.CW.CurrentCostEstimate()
		if err != nil {
			logging.Error("Error getting current cost estimate", "error", err)
			ctx.Status(500)
			return
		}
		currentPlayerEstimate = &_currentEstimate.Players
		isLive = true
	}

	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		ctx.Status(500)
		return
	}

	response, maxDuration := computeTotalChartResponse(ctx, app, players, billedPlayerTime, currentPlayerEstimate)
	utils.Render(ctx, 200, playtime_view.TotalChart(fmt.Sprintf("%d", year), isLive, fmt.Sprintf("%d", year), response, maxDuration, false, nil))
}
