package widgets

import (
	"context"
	"mc_manager/model"
	"net/http"

	widgets_view "mc_manager/web/views/components/widgets"
	utils "mc_manager/web/web_utils"

	"github.com/gin-gonic/gin"
)

func getServerStatusHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		isOnline := app.SW.IsServerOnline()
		isReady := app.SW.IsServerReady()

		serverState := model.SERVER_STATUS_OFFLINE
		if isOnline && !isReady {
			serverState = model.SERVER_STATUS_STARTING
		} else if isOnline && isReady {
			serverState = model.SERVER_STATUS_ONLINE
		}

		utils.Render(ctx, http.StatusOK, widgets_view.ServerStatus(serverState))
	}
}
