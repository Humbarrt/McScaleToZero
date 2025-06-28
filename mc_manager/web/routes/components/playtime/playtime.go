package playtime

import (
	"context"
	"mc_manager/logging"
	"mc_manager/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.RouterGroup, app *model.App) {
	bills := route.Group("/playtime")
	{
		bills.GET("/selector", getYearSelector(app))
		bills.GET("/selector/:year", getMonthSelector(app))
		bills.GET("/selector/:year/:month", getDaySelector(app))
		bills.GET("/:chartId", getChart(app))
	}
}

func getChart(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		chartIdString := ctx.Param("chartId")

		isTotal := strings.ToLower(chartIdString) == "total"
		if isTotal {
			renderTotalPlaytime(ctx, app, chartIdString)
			return
		}

		if len(chartIdString) > 8 || len(chartIdString) < 4 {
			logging.Error("Invalid chart ID format", "chartId", chartIdString)
			ctx.JSON(400, gin.H{"error": "Invalid chart ID format"})
			return
		}

		year, err := strconv.Atoi(chartIdString[:4])
		if err != nil {
			logging.Error("Invalid year format", "error", err, "chartId", chartIdString)
			ctx.JSON(400, gin.H{"error": "Invalid year format"})
			return
		}

		var month *int
		if len(chartIdString) > 4 {
			_month, err := strconv.Atoi(chartIdString[4:6])
			if err != nil {
				logging.Error("Invalid month format", "error", err, "chartId", chartIdString)
				ctx.JSON(400, gin.H{"error": "Invalid month format"})
				return
			}
			month = &_month
		}
		if month != nil && (*month < 1 || *month > 12) {
			logging.Error("Invalid month value", "month", *month, "chartId", chartIdString)
			ctx.JSON(400, gin.H{"error": "Invalid month value"})
			return
		}

		var day *int
		if len(chartIdString) > 6 {
			_day, err := strconv.Atoi(chartIdString[6:8])
			if err != nil {
				logging.Error("Invalid day format", "error", err, "chartId", chartIdString)
				ctx.JSON(400, gin.H{"error": "Invalid day format"})
				return
			}
			day = &_day
		}

		timezone, usingLocalTime, err := getTimezoneFromContext(ctx)
		if err != nil {
			logging.Error("Error getting timezone from context", "error", err)
			return
		}
		logging.Info("Using timezone", "timezone", timezone, "usingLocalTime", usingLocalTime)

		if month == nil && day == nil {
			logging.Info("Rendering year playtime", "year", year)
			renderYearPlaytime(ctx, app, year)
		} else if month != nil && day == nil {
			renderMonthPlaytime(ctx, app, year, *month, usingLocalTime)
		} else if month != nil && day != nil {
			renderDayPlaytime(ctx, app, year, *month, *day, timezone, usingLocalTime)
		} else {
			logging.Error("Invalid chart ID format", "chartId", chartIdString)
			ctx.JSON(400, gin.H{"error": "Invalid chart ID format"})
		}
	}
}
