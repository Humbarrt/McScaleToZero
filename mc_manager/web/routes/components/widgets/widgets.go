package widgets

import (
	"mc_manager/model"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.RouterGroup, app *model.App) {
	widgets := route.Group("/widgets")
	{
		widgets.GET("/server-status", getServerStatusHandler(app))
		widgets.GET("/active-players", getActivePlayersHandler(app))
		widgets.GET("/total-uptime", getTotalUptimeHandler(app))
		widgets.GET("/total-cost", getTotalCostHandler(app))
	}
}
