package middleware

import (
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	auth_view "mc_analyser/web/views/auth"
	utils "mc_analyser/web/web_utils"
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
			return app.Config.SIGNING_KEY, nil
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
		_, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logging.Error("Invalid claims")
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			return
		}

		//log.Println(claims)
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
