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

func getActivePlayersHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		maxPlayers := app.Config.GetMaxPlayers()
		playersEvents, err := queries.GetLastestPlayersEvents(app.Db)
		if err != nil {
			logging.Error("Error getting players events", "error", err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		if playersEvents == nil {
			logging.Error("Error getting players events", "error", "playersEvents is nil")
			ctx.Status(http.StatusInternalServerError)
			return
		}

		activePlayers := 0
		for _, event := range playersEvents {
			if event.IsOnline {
				activePlayers++
			}
		}
		utils.Render(ctx, http.StatusOK, widgets_view.ActivePlayers(activePlayers, maxPlayers))
	}
}
