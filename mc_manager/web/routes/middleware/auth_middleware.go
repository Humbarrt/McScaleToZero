package middleware

import (
	"fmt"
	"mc_manager/logging"
	model "mc_managermodel"
	auth_view "mc_managerweb/views/auth"
	utils "mc_managerweb/web_utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func IsAuthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("auth", false)

		cookie, err := ctx.Cookie("session")
		if err != nil {
			ctx.Next()
			return
		}

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return app.JWTSecret, nil
		})

		if err != nil {
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			ctx.Next()
			return
		}

		if !token.Valid {
			logging.Error("Invalid token")
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			ctx.Next()
			return
		}

		//claims, ok := token.Claims.(jwt.MapClaims)
		//TODO: check claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logging.Error("Invalid claims")
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			return
		}

		logging.Debug("Authenticated", "claims", claims)
		ctx.Set("auth", true)
		ctx.Next()
	}
}

func BlockIfUnauthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, _ := ctx.Get("auth")
		if auth == false {
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func RenderAuthPageIfUnauthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, _ := ctx.Get("auth")
		if auth == false {
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.Index())
			return
		}
		ctx.Next()
	}
}
