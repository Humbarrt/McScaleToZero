package player

import (
	"mc_manager/model"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.RouterGroup, app *model.App) {
	widgets := route.Group("/player")
	{
		widgets.GET("/sessions", getPlayerSessionsHandler(app))
		widgets.DELETE("/sessions/:sessionId", deleteSessionHandler(app))
	}
}
