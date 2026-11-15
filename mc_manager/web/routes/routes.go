package routes

import (
	"fmt"
	"io/fs"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes/auth"
	"mc_manager/web/routes/components"
	"mc_manager/web/routes/events"
	"mc_manager/web/routes/index"
	"mc_manager/web/routes/middleware"
	"mc_manager/web/routes/player"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes(app *model.App) error {

	app.Router.Use(gin.Recovery())
	app.Router.Use(middleware.SlogMiddleware(logging.Logger))

	//CDN
	cdn, err := fs.Sub(app.FS, "web/static")
	if err != nil {
		return err
	}
	app.Router.StaticFS("/static", http.FS(cdn))
	logging.Info("Static files loaded")

	//Check authentication for routes
	app.Router.Use(middleware.IsAuthenticated(app))

	index.Routes(app.Router, app)
	auth.Routes(app.Router, app)
	components.Routes(app.Router, app)
	events.Routes(app.Router, app)
	player.Routes(app.Router, app)

	invoiceService := app.Router.Group("/is")
	invoiceService.Use(middleware.BlockIfUnauthenticated())
	app.IS.Routes(invoiceService, app)
	return app.Router.Run(fmt.Sprintf(":%d", app.Config.GetDashboardPort()))
}
