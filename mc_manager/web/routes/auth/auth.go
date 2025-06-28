package auth

import (
	"context"

	"mc_manager/model"
	"mc_manager/web/routes/middleware"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"

	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	auth := route.Group("/auth")
	{
		auth.GET("/", authPageHandler(app))
		auth.GET("/signin", signInPageHandler(app))
		auth.POST("/signin", signInHandler(app))

		auth.GET("/reset", getResetPasswordPageHandler(app))
		auth.POST("/reset", resetPasswordHandler(app))
		auth.GET("/forgot", getForgotPasswordPageHandler(app))

		auth.GET("/logout", middleware.BlockIfUnauthenticated(), getLogoutHandler(app))
	}
}

func authPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()
		utils.Render(ctx, http.StatusOK, auth_view.Index())
	}
}
