package widgets

import (
	"context"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"net/http"

	widgets_view "mc_manager/web/views/components/widgets"
	utils "mc_manager/web/web_utils"

	"github.com/gin-gonic/gin"
)

func getTotalCostHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		billedTotalCost, err := queries.GetBilledTotalCost(app.Db)
		if err != nil {
			logging.Error("Error getting billed total cost", "error", err)
			utils.Render(ctx, http.StatusOK, widgets_view.TotalCost(nil))
			return
		}

		currentEstimate, err := app.CW.CurrentCostEstimate()
		if err != nil {
			logging.Error("Error getting current cost estimate", "error", err)
			utils.Render(ctx, http.StatusOK, widgets_view.TotalCost(nil))
			return
		}

		totalCost := billedTotalCost.Add(currentEstimate.TotalCost)
		utils.Render(ctx, http.StatusOK, widgets_view.TotalCost(&totalCost))
	}
}
