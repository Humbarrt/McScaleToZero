package bills

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

		if len(years) < 2 {
			ctx.Redirect(http.StatusFound, fmt.Sprintf("/components/bills/selector/%d", currentYear))
			return
		}

		selectorId := "billsSelector"
		options := []model.SelectorOption{}
		for _, year := range years {
			options = append(options, model.SelectorOption{
				Label:  fmt.Sprintf("%d", year),
				Url:    fmt.Sprintf("/components/bills/selector/%d", year),
				Target: fmt.Sprintf("#%sHolder", selectorId),
				Swap:   "outerHTML",
			})
		}

		utils.Render(ctx, http.StatusOK, selector_view.Selector(selectorId, options, "Select a year", "BillsHolder", ""))
	}
}

func getMonthSelector(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		logging.Info("Retrieving month selector for year")
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

		years := []int{}
		thisYearsMonths := []model.YearMonth{}
		for _, month := range months {
			if len(years) == 0 || years[len(years)-1] != month.Year() {
				years = append(years, month.Year())
			}
			if month.Year() == year {
				thisYearsMonths = append(thisYearsMonths, month)
			}
		}
		currentYear := time.Now().Year()
		if len(years) == 0 || years[len(years)-1] != currentYear {
			years = append(years, currentYear)
		}

		selectorId := "billsSelector"
		options := []model.SelectorOption{}
		if len(years) > 1 {
			options = append(options, model.SelectorOption{
				Label:  "..",
				Url:    "/components/bills/selector",
				Target: fmt.Sprintf("#%sHolder", selectorId),
				Swap:   "outerHTML",
			})
		}

		for i := range thisYearsMonths {
			options = append(options, model.SelectorOption{
				Label:  model.MonthsLabels[thisYearsMonths[i].Month()-1],
				Url:    fmt.Sprintf("/components/bills/%s", thisYearsMonths[i].String()),
				Target: "#BillsHolder",
				Swap:   "outerHTML",
			})
		}

		if time.Now().Year() == year {
			options = append(options, model.SelectorOption{
				Label:  model.MonthsLabels[time.Now().Month()-1],
				Url:    "/components/bills/latest",
				Target: "#BillsHolder",
				Swap:   "outerHTML",
			})
		}

		utils.Render(ctx, http.StatusOK, selector_view.Selector(selectorId, options, "Select a month", "BillsHolder", ""))
	}
}
