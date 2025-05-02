package auth

import (
	"context"
	"fmt"
	"time"

	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TODO: Login session
// TODO: Login Timeout
func Routes(route *gin.Engine, app *model.App) {
	auth := route.Group("/auth")
	{
		auth.GET("/", authPageHandler(app))
		auth.POST("/signin", signInHandler(app))
	}
}

func authPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()
		utils.Render(ctx, http.StatusOK, auth_view.Index())
	}
}

func signInHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var json struct {
			Username string `json:"username"`
		}

		if err := ctx.ShouldBindJSON(&json); err != nil {
			logging.Error("error", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "general", "message": "Invalid Request"})
			return
		}

		username := json.Username
		if username == "" {
			logging.Error("error, username is empty")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "username", "message": "Username is empty"})
			return
		}

		logging.Debug("Auth request from", "username", username)
		user, err := queries.GetPlayerByAlias(app.Db, username)
		if err != nil || user == nil {
			logging.Error("error", "error", err, "user", user)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "general", "message": "Internal Error, please try again"})
			return
		}

		isCheckingIP := app.Config.GetIPCheck()
		if isCheckingIP && (user.LastIPV4 == nil || *user.LastIPV4 != ctx.ClientIP()) {
			logging.Error("error, ip address does not match", "ip", ctx.ClientIP(), "username", user.Aliases[0])
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ip", "message": "Please join the MC server from the same IP address"})
			return
		}

		claims := model.AuthClaims{
			McUUID: fmt.Sprintf("%d", user.Id),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(365 * 24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(app.JWTSecret)
		if err != nil {
			logging.Error("error", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "general", "message": "Internal Error, please try again"})
			return
		}

		ctx.SetSameSite(http.SameSiteStrictMode)
		ctx.SetCookie("session", tokenString, 3600*24*365*100, "/", "", false, true)
		ctx.Status(http.StatusOK)
	}
}
