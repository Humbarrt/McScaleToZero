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

func renderMonthPlaytime(ctx *gin.Context, app *model.App, year int, month int, usingLocalTime bool) {
	start := year*100 + month
	end := start + 1
	billedPlayerTime, err := queries.GetBilledPlayerTimeForTimespan(app.Db, start, end)
	if err != nil {
		logging.Error("Error getting billed player time", "error", err)
		ctx.Status(500)
		return
	}

	isLive := false
	var currentPlayerEstimate *[]model.PlayerCostEstimate
	if year == time.Now().Year() && month == int(time.Now().Month()) {
		_currentEstimate, err := app.CW.CurrentCostEstimate()
		if err != nil {
			logging.Error("Error getting current cost estimate", "error", err)
			ctx.Status(500)
			return
		}
		currentPlayerEstimate = &_currentEstimate.Players
		isLive = true
	}

	ym, err := model.NewYearMonthFromInt(year, month)
	if err != nil {
		logging.Error("Error creating YearMonth", "error", err)
		ctx.JSON(500, gin.H{"error": "Error parsing year and month"})
		return
	}

	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		ctx.Status(500)
		return
	}

	playersMap := make(map[int64]*model.Player)
	for _, player := range players {
		playersMap[player.Id] = player
	}

	events, err := queries.GetPlayersEvents(app.Db, ym.Timeframe())
	if err != nil {
		logging.Error("Error getting player events", "error", err)
		ctx.Status(500)
		return
	}

	eventList := model.EventListResponse{
		Players: playersMap,
		Events:  events,
	}

	response, maxDuration := computeTotalChartResponse(ctx, app, players, billedPlayerTime, currentPlayerEstimate)
	utils.Render(ctx, 200, playtime_view.TotalChart(fmt.Sprintf("%d%02d", year, month), isLive, fmt.Sprintf("%d %s", year, model.MonthsLabels[month-1]), response, maxDuration, usingLocalTime, &eventList))
}
