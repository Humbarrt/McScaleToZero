package index

import (
	"context"
	"mc_manager/model"
	"mc_managerweb/routes/middleware"
	"mc_managerweb/views/index"
	utils "mc_managerweb/web_utils"

	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	index := route.Group("/")
	{
		index.GET("/", middleware.RenderAuthPageIfUnauthenticated(app), indexPageHandler(app))
	}
}

func indexPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()
		utils.Render(ctx, http.StatusOK, index.Index())
	}
}
