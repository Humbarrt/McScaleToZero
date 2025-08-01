package auth

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getForgotPasswordPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		playerIdString := ctx.Query("playerId")
		playerId, err := strconv.ParseInt(playerIdString, 10, 64)
		if err != nil {
			logging.Error("getForgotPasswordPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("invalid player ID")))
			return
		}

		player, err := queries.GetPlayerById(app.Db, playerId)
		if err != nil || player == nil {
			logging.Error("SignInPageHandler", "error", err, "player", player)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("player not found")))
			return
		}

		session, err := app.SM.CreatePasswordResetSession(player.Id)
		if err != nil || session == nil {
			logging.Error("SignInPageHandler", "error", err)
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("internal error, please try again")))
			return
		}

		utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(player, session, nil))
	}
}
