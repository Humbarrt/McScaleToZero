package routes

import (
	"fmt"
	"io/fs"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes/auth"
	"mc_manager/web/routes/index"
	"mc_manager/web/routes/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes(app *model.App) error {

	app.Router.Use(gin.Recovery())
	app.Router.Use(middleware.SlogMiddleware(logging.Logger))
	app.Router.Use(middleware.IsAuthenticated(app))

	//CDN
	cdn, err := fs.Sub(app.FS, "web/static")
	if err != nil {
		return err
	}
	app.Router.StaticFS("/static", http.FS(cdn))

	logging.Info("Static files loaded")

	//Routes
	index.Routes(app.Router, app)
	auth.Routes(app.Router, app)
	app.Router.Run(fmt.Sprintf(":%d", app.Config.GetDashboardPort()))
	return nil
}
