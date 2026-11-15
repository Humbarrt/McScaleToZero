package index

import (
	"context"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes/middleware"
	"mc_manager/web/views/index"
	utils "mc_manager/web/web_utils"
	"runtime"

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

		logging.Debug("Runtime Status", "Goroutines", runtime.NumGoroutine())

		isAdmin, err := utils.IsAdmin(ctx)
		if err != nil {
			logging.Error("Error checking admin status", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		utils.Render(ctx, http.StatusOK, index.Index(isAdmin, app.IS))
	}
}
