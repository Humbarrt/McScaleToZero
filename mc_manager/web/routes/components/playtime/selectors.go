package playtime

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	selector_view "mc_manager/web/views/components/selector"
	utils "mc_manager/web/web_utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func getYearSelector(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		selectorId := "playtimeSelector"
		options := []model.SelectorOption{}
		options = append(options, model.SelectorOption{
			Label:  "Total",
			Url:    "/components/playtime/total",
			Target: "#PlaytimeHolder",
			Swap:   "outerHTML",
		})

		months, err := queries.GetBillingMonths(app.Db)
		if err != nil {
			logging.Error("Error retrieving billing months", "error", err)
			ctx.JSON(500, gin.H{"error": "Error retrieving billing months"})
			return
		}

		years := []int{}
		for _, month := range months {
			if len(years) == 0 || years[len(years)-1] != month.Year() {
				years = append(years, month.Year())
			}
		}
		currentYear := time.Now().Year()
		if len(years) == 0 || years[len(years)-1] != currentYear {
			years = append(years, currentYear)
		}

		for _, year := range years {
			options = append(options, model.SelectorOption{
				Label:  fmt.Sprintf("%d", year),
				Url:    fmt.Sprintf("/components/playtime/selector/%d", year),
				Target: fmt.Sprintf("#%sHolder", selectorId),
				Swap:   "outerHTML",
			})
		}

		utils.Render(ctx, http.StatusOK, selector_view.Selector(selectorId, options, "Select a year", "PlaytimeHolder", ""))
	}
}

func getMonthSelector(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		yearString := ctx.Param("year")
		if yearString == "" {
			ctx.JSON(500, gin.H{"error": "Year parameter is required"})
			return
		}

		year, err := strconv.Atoi(yearString)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Invalid year format"})
			return
		}

		months, err := queries.GetBillingMonths(app.Db)
		if err != nil {
			logging.Error("Error retrieving billing months", "error", err)
			ctx.JSON(500, gin.H{"error": "Error retrieving billing months"})
			return
		}

		thisYearsMonths := []model.YearMonth{}
		for _, month := range months {
			if month.Year() == year {
				thisYearsMonths = append(thisYearsMonths, month)
			}
		}

		if time.Now().Year() == year {
			thisYearsMonths = append(thisYearsMonths, model.NewYearMonthFromTime(time.Now()))
		}

		selectorId := "playtimeSelector"
		options := []model.SelectorOption{}
		options = append(options, model.SelectorOption{
			Label:  "..",
			Url:    "/components/playtime/selector",
			Target: fmt.Sprintf("#%sHolder", selectorId),
			Swap:   "outerHTML",
		})

		options = append(options, model.SelectorOption{
			Label:  fmt.Sprintf("%d", year),
			Url:    fmt.Sprintf("/components/playtime/%d", year),
			Target: "#PlaytimeHolder",
			Swap:   "outerHTML",
			Style:  "padding-left: 0rem; padding-right: 0rem;",
		})

		for _, month := range thisYearsMonths {
			options = append(options, model.SelectorOption{
				Label:  model.MonthsLabels[month.Month()-1],
				Url:    fmt.Sprintf("/components/playtime/selector/%d/%d", year, month.Month()),
				Target: fmt.Sprintf("#%sHolder", selectorId),
				Swap:   "outerHTML",
			})
		}

		utils.Render(ctx, http.StatusOK, selector_view.Selector(selectorId, options, "Select a month", "PlaytimeHolder", ""))
	}
}

func getDaySelector(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		yearString := ctx.Param("year")
		if yearString == "" {
			ctx.JSON(500, gin.H{"error": "Year parameter is required"})
			return
		}
		year, err := strconv.Atoi(yearString)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Invalid year format"})
			return
		}

		monthString := ctx.Param("month")
		if monthString == "" {
			ctx.JSON(500, gin.H{"error": "Month parameter is required"})
			return
		}
		month, err := strconv.Atoi(monthString)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Invalid month format"})
			return
		}

		if month < 1 || month > 12 {
			logging.Error("Invalid month value", "month", month)
			ctx.JSON(400, gin.H{"error": "Invalid month value"})
			return
		}

		timezone, _, err := getTimezoneFromContext(ctx)
		if err != nil {
			logging.Error("Error getting timezone from context", "error", err)
			ctx.JSON(500, gin.H{"error": "Error getting timezone"})
			return
		}

		ym, err := model.NewYearMonthFromInt(year, month)
		if err != nil {
			logging.Error("Error creating YearMonth", "error", err)
			ctx.JSON(500, gin.H{"error": "Error parsing year and month"})
			return
		}
		tf := model.Timeframe{
			Start: ym.StartTime().In(timezone),
			End:   ym.EndTime().In(timezone),
		}

		events, err := queries.GetPlayersEvents(app.Db, tf)
		if err != nil {
			logging.Error("Error getting player events", "error", err)
			ctx.Status(500)
			return
		}

		days := make([]bool, ym.EndTime().Day())
		for _, event := range events {
			days[event.Timestamp.Day()-1] = true
		}

		selectorId := "playtimeSelector"
		options := []model.SelectorOption{}
		options = append(options, model.SelectorOption{
			Label:  "..",
			Url:    fmt.Sprintf("/components/playtime/selector/%d", year),
			Target: fmt.Sprintf("#%sHolder", selectorId),
			Swap:   "outerHTML",
		})

		options = append(options, model.SelectorOption{
			Label:  model.MonthsLabels[month-1],
			Url:    fmt.Sprintf("/components/playtime/%d%02d", year, month),
			Target: "#PlaytimeHolder",
			Swap:   "outerHTML",
			Style:  "padding-left: 0rem; padding-right: 0rem;",
		})

		for day, active := range days {
			options = append(options, model.SelectorOption{
				Label:    fmt.Sprintf("%02d", day+1),
				Url:      fmt.Sprintf("/components/playtime/%d%02d%02d", year, month, day+1),
				Target:   "#PlaytimeHolder",
				Swap:     "outerHTML",
				Disabled: !active,
			})
		}

		utils.Render(ctx, http.StatusOK, selector_view.Selector(selectorId, options, "Select a day", "PlaytimeHolder", "grid-template-columns: repeat(5, minmax(0, 1fr));"))
	}
}
