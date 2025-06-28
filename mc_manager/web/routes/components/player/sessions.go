package player

import (
	"context"
	"mc_manager/model"
	player_view "mc_manager/web/views/components/player"
	utils "mc_manager/web/web_utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPlayerSessionsHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		playerIdAny, exists := ctx.Get("playerId")
		if !exists || playerIdAny == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		playerId, ok := playerIdAny.(int64)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player ID"})
			return
		}

		sessionIdAny, exists := ctx.Get("sessionId")
		if !exists || sessionIdAny == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		sessionId, ok := sessionIdAny.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session ID"})
			return
		}

		playerSessions, err := app.SM.GetPlayerSessions(playerId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve player sessions"})
			return
		}

		utils.Render(ctx, http.StatusOK, player_view.PlayerSessions(playerSessions, sessionId))
	}
}

func deleteSessionHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		playerId, exists := ctx.Get("playerId")
		if !exists || playerId == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		playerIdInt64, ok := playerId.(int64)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player ID"})
			return
		}

		sessionId := ctx.Param("sessionId")
		if sessionId == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}

		err := app.SM.RemoveUserSession(playerIdInt64, sessionId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		getPlayerSessionsHandler(app)(ctx)
	}
}
