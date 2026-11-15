package settings

import (
	"mc_manager/model"
	"mc_manager/web/routes/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.RouterGroup, app *model.App) {
	settings := route.Group("/settings")
	{
		settings.GET("/server", getServerConfigHandler(app))
		settings.POST("/server", middleware.BlockIfNotAdmin(), postServerConfigHandler(app))
		settings.POST("/whitelist", middleware.BlockIfNotAdmin(), postWhitelistUpdateHandler(app))

		settings.GET("/billing", getBillingConfigHandler(app))
		settings.POST("/billing", middleware.BlockIfNotAdmin(), postBillingConfigHandler(app))

		settings.GET("/info", getServerInfoHandler(app))
	}
}
