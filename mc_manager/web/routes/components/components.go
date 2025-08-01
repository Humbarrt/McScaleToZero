package components

import (
	"mc_manager/model"
	"mc_manager/web/routes/components/bills"
	"mc_manager/web/routes/components/player"
	"mc_manager/web/routes/components/playtime"
	"mc_manager/web/routes/components/settings"
	"mc_manager/web/routes/components/widgets"
	"mc_manager/web/routes/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	components := route.Group("/components")
	components.Use(middleware.BlockIfUnauthenticated())

	widgets.Routes(components, app)
	settings.Routes(components, app)
	playtime.Routes(components, app)
	bills.Routes(components, app)
	player.Routes(components, app)
}
