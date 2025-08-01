package player

import (
	"fmt"
	"io"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	player := route.Group("/player")
	{
		player.GET("/:playerId/avatar/head", middleware.BlockIfUnauthenticated(), getAvatarHeadHandler(app))
	}
}

func getAvatarHeadHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		playerIdString := ctx.Param("playerId")
		if playerIdString == "" {
			ctx.JSON(500, gin.H{"error": "Player ID is required"})
			return
		}

		playerId, err := strconv.ParseInt(playerIdString, 10, 64)
		if err != nil {
			logging.Error("Error parsing player ID", "error", err)
			ctx.JSON(500, gin.H{"error": "Invalid player ID format"})
			return
		}

		avatarHead, lastDownloaded, err := queries.GetPlayerAvatarHead(app.Db, playerId)
		if err != nil {
			logging.Error("Error getting player avatar head", "error", err)
			ctx.String(http.StatusInternalServerError, "Error getting avatar head")
			return
		}

		if avatarHead != nil || len(avatarHead) > 0 {
			ctx.Data(http.StatusOK, "image/png", avatarHead)
			return
		}

		if lastDownloaded != nil && time.Since(*lastDownloaded) < 24*time.Hour {
			ctx.JSON(http.StatusOK, gin.H{"error": "wait 24 hours before trying another download"})
			return
		}

		player, err := queries.GetPlayerById(app.Db, playerId)
		if err != nil {
			logging.Error("Error getting player", "error", err)
			ctx.String(http.StatusInternalServerError, "Error getting player")
			return
		}

		if len(player.Aliases) == 0 {
			ctx.String(http.StatusNotFound, "Player has no aliases")
			return
		}

		err = queries.UpdatePlayerAvatarHead(app.Db, player.Id, nil, time.Now())
		if err != nil {
			logging.Error("Error updating avatar head", "error", err)
			ctx.String(http.StatusInternalServerError, "Error updating avatar head")
			return
		}

		url := fmt.Sprintf("https://vzge.me/head/%s.png?no=shadow&y=70", player.Aliases[0])
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logging.Error("Error creating request", "error", err)
			ctx.String(http.StatusInternalServerError, "Error creating request")
			return
		}
		req.Header.Set("User-Agent", "McScaleToZero/1.0 (+https://github.com/Humbarrt/McScaleToZero)")
		resp, err := client.Do(req)
		if err != nil {
			logging.Error("Error downloading avatar head", "error", err)
			ctx.String(http.StatusInternalServerError, "Error downloading avatar head")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logging.Error("Avatar head not found", "status", resp.StatusCode, "url", url)
			ctx.String(http.StatusNotFound, "Avatar head not found")
			return
		}

		avatarHead, err = io.ReadAll(resp.Body)
		if err != nil {
			logging.Error("Error reading avatar head", "error", err)
			ctx.String(http.StatusInternalServerError, "Error reading avatar head")
			return
		}

		err = queries.UpdatePlayerAvatarHead(app.Db, player.Id, avatarHead, time.Now())
		if err != nil {
			logging.Error("Error updating avatar head", "error", err)
			ctx.String(http.StatusInternalServerError, "Error updating avatar head")
			return
		}

		ctx.Data(http.StatusOK, "image/png", avatarHead)
	}
}
