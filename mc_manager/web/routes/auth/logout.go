package auth

import (
	"context"
	"fmt"
	"mc_manager/logging"
	"mc_manager/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getLogoutHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		playerIdAny, ok := ctx.Get("playerId")
		if !ok || playerIdAny == nil {
			logging.Error("Player ID not found in context")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		playerId, ok := playerIdAny.(int64)
		if !ok {
			logging.Error("Player ID is not an int64")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		sessionIdAny, ok := ctx.Get("sessionId")
		if !ok || sessionIdAny == nil {
			logging.Error("Session ID not found in context")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		sessionId, ok := sessionIdAny.(string)
		if !ok {
			logging.Error("Session ID is not a string")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		app.SM.RemoveUserSession(playerId, fmt.Sprintf("%v", sessionId))

		ctx.Header("HX-Refresh", "true")
		ctx.Status(http.StatusOK)
	}
}
