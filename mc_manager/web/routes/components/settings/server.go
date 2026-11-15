package settings

import (
	"context"
	"mc_manager/config"
	"mc_manager/logging"
	"mc_manager/model"
	"net/http"
	"strconv"
	"time"

	settings_view "mc_manager/web/views/components/settings"
	utils "mc_manager/web/web_utils"

	"github.com/gin-gonic/gin"
)

func getServerConfigHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		isAdmin, err := utils.IsAdmin(ctx)
		if err != nil {
			logging.Error("Error checking admin status", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		hasSaved := ctx.Query("serverConfigSuccessIcon") == "true"

		serverPort := app.Config.GetMCServerPort()
		backupsToKeep := app.Config.GetBackupsToKeep()
		replaceMOTD := app.Config.GetReplaceMOTD()
		offlineMOTD := app.Config.GetMCOfflineMOTD()
		onlineMOTD := app.Config.GetMCOnlineMOTD()
		offlineImage := app.Config.GetMCOfflineImage()
		onlineImage := app.Config.GetMCOnlineImage()

		utils.Render(ctx, http.StatusOK, settings_view.ServerConfig(isAdmin, serverPort, backupsToKeep, replaceMOTD, offlineMOTD, onlineMOTD, offlineImage, onlineImage, hasSaved, app.SW.IsUpdatingWhitelist(), app.SW.IsServerReady()))
	}
}

func postServerConfigHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			ServerPort    string `form:"serverPort"`
			BackupsToKeep string `form:"backupsToKeep"`
			ReplaceMOTD   string `form:"replaceMOTD"`
			OfflineMOTD   string `form:"offlineMOTD"`
			OnlineMOTD    string `form:"onlineMOTD"`
			OfflineImage  string `form:"offlineImage"`
			OnlineImage   string `form:"onlineImage"`
		}
		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("Error binding form", "error", err)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid form")
			return
		}

		if form.ReplaceMOTD != "true" && form.ReplaceMOTD != "false" {
			logging.Error("Invalid replaceMOTD value", "replaceMOTD", form.ReplaceMOTD)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid replaceMOTD value")
			return
		}
		replaceMOTD := form.ReplaceMOTD == "true"

		var offlineImage *string
		if form.OfflineImage != "" && form.OfflineImage != "#" {
			offlineImage = &form.OfflineImage
			err := config.ValidateImage(offlineImage)
			if err != nil {
				logging.Error("Error validating offline image", "error", err)
				utils.RenderError(ctx, http.StatusBadRequest, "Invalid offline icon")
				return
			}
		}

		var onlineImage *string
		if form.OnlineImage != "" && form.OnlineImage != "#" {
			onlineImage = &form.OnlineImage
			err := config.ValidateImage(onlineImage)
			if err != nil {
				logging.Error("Error validating online image", "error", err)
				utils.RenderError(ctx, http.StatusBadRequest, "Invalid online icon")
				return
			}
		}

		serverPort, err := strconv.Atoi(form.ServerPort)
		if err != nil {
			logging.Error("Error converting serverPort to int", "error", err)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid serverPort value")
			return
		}
		if serverPort < 0 || serverPort > 65535 {
			logging.Error("Invalid serverPort value", "serverPort", serverPort)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid serverPort value")
			return
		}

		backupsToKeep, err := strconv.Atoi(form.BackupsToKeep)
		if err != nil {
			logging.Error("Error converting backupsToKeep to int", "error", err)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid backupsToKeep value")
			return
		}
		if backupsToKeep < 0 || backupsToKeep > 15 {
			logging.Error("Invalid backupsToKeep value", "backupsToKeep", backupsToKeep)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid backupsToKeep value")
			return
		}

		offlineMOTD := []byte(form.OfflineMOTD)
		if len(offlineMOTD) > 1024 {
			logging.Error("Offline MOTD is too long", "length", len(offlineMOTD))
			utils.RenderError(ctx, http.StatusBadRequest, "Offline MOTD is too long")
			return
		}

		onlineMOTD := []byte(form.OnlineMOTD)
		if len(onlineMOTD) > 1024 {
			logging.Error("Online MOTD is too long", "length", len(onlineMOTD))
			utils.RenderError(ctx, http.StatusBadRequest, "Online MOTD is too long")
			return
		}

		err = app.Config.SetMCServerPort(serverPort)
		if err != nil {
			logging.Error("Error setting server port", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting server port")
			return
		}
		err = app.Config.SetBackupsToKeep(backupsToKeep)
		if err != nil {
			logging.Error("Error setting backups to keep", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting backups to keep")
			return
		}
		err = app.Config.SetReplaceMOTD(replaceMOTD)
		if err != nil {
			logging.Error("Error setting replace MOTD", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting replace MOTD")
			return
		}
		err = app.Config.SetMCOfflineMOTD(offlineMOTD)
		if err != nil {
			logging.Error("Error setting offline MOTD", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting offline MOTD")
			return
		}
		err = app.Config.SetMCOnlineMOTD(onlineMOTD)
		if err != nil {
			logging.Error("Error setting online MOTD", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting online MOTD")
			return
		}
		err = app.Config.SetMCOfflineImage(offlineImage)
		if err != nil {
			logging.Error("Error setting offline image", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting offline icon")
			return
		}
		err = app.Config.SetMCOnlineImage(onlineImage)
		if err != nil {
			logging.Error("Error setting online image", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting online icon")
			return
		}

		newServerPort := app.Config.GetMCServerPort()
		newBackupsToKeep := app.Config.GetBackupsToKeep()
		newReplaceMOTD := app.Config.GetReplaceMOTD()
		newOfflineMOTD := app.Config.GetMCOfflineMOTD()
		newOnlineMOTD := app.Config.GetMCOnlineMOTD()
		newOfflineImage := app.Config.GetMCOfflineImage()
		newOnlineImage := app.Config.GetMCOnlineImage()

		utils.Render(ctx, http.StatusOK, settings_view.ServerConfig(true, newServerPort, newBackupsToKeep, newReplaceMOTD, newOfflineMOTD, newOnlineMOTD, newOfflineImage, newOnlineImage, true, app.SW.IsUpdatingWhitelist(), app.SW.IsServerReady()))

		go func() {
			time.Sleep(100 * time.Millisecond)
			app.EB.Broadcast(model.EB_EVENT_SERVER_CONFIG_UPDATE)
		}()
	}
}

func postWhitelistUpdateHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		go func() {
			err := app.SW.UpdateWhitelistAndAdmins()
			if err != nil {
				logging.Error("Error updating whitelist", "error", err)
				utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
				return
			}
		}()

		time.Sleep(100 * time.Millisecond) // Give the goroutine a moment to start

		//then render getServerConfigHandler
		getServerConfigHandler(app)(ctx)
	}
}
