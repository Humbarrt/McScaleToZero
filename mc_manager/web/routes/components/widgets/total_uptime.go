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

func getTotalUptimeHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		billedServeruptime, err := queries.GetBilledServerUptime(app.Db)
		if err != nil {
			logging.Error("Error getting billed server uptime", "error", err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		currentEstimate, err := app.CW.CurrentCostEstimate()
		if err != nil {
			logging.Error("Error getting current cost estimate", "error", err)
			utils.Render(ctx, http.StatusOK, widgets_view.TotalUptime(nil))
			return
		}

		totalUptime := billedServeruptime + currentEstimate.ServerUptime
		utils.Render(ctx, http.StatusOK, widgets_view.TotalUptime(&totalUptime))
	}
}
