package events

import (
	"mc_manager/model"
	"mc_manager/web/routes/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	events := route.Group("/events")
	{
		events.GET("/", middleware.BlockIfUnauthenticated(), app.EB.Middleware())
	}
}
