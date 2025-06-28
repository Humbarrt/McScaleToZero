package settings

import (
	"context"
	"mc_manager/model"
	"net/http"

	settings_view "mc_manager/web/views/components/settings"
	utils "mc_manager/web/web_utils"

	"github.com/gin-gonic/gin"
)

func getServerInfoHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		lastMOTD := app.Config.GetLastMOTD()
		lastImage := app.Config.GetLastImage()
		protocolVersion := app.Config.GetProtocolVersion()
		maxPlayers := app.Config.GetMaxPlayers()

		utils.Render(ctx, http.StatusOK, settings_view.ServerInfo(lastMOTD, lastImage, protocolVersion, maxPlayers))
	}
}
