package middleware

import (
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func IsAuthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("playerId", nil)
		ctx.Set("auth", false)
		ctx.Set("isAdmin", false)
		ctx.Set("sessionId", nil)

		cookie, err := ctx.Cookie("session")
		if err != nil {
			logging.Debug("No session cookie found", "error", err)
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			ctx.Next()
			return
		}

		claims := &model.AuthClaims{}
		token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return app.JWTSecret, nil
		})

		if err != nil {
			logging.Error("Error parsing JWT", "error", err)
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

		_, err = app.SM.ExtendSession(claims.PlayerId, claims.SessionId)
		if err != nil {
			logging.Error("Error extending session", "error", err)
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			ctx.Next()
			return
		}

		player, err := queries.GetPlayerById(app.Db, claims.PlayerId)
		if err != nil || player == nil {
			logging.Error("Error getting player", "error", err, "player", player)
			ctx.SetCookie("session", "", 0, "/", "", false, true)
			ctx.Next()
			return
		}

		if player.IsAdmin {
			ctx.Set("isAdmin", true)
		}

		ctx.Set("playerId", claims.PlayerId)
		ctx.Set("auth", true)
		ctx.Set("sessionId", claims.SessionId)
		ctx.Next()
	}
}

func BlockIfUnauthenticated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if !exists || auth == false {
			utils.RenderError(ctx, http.StatusForbidden, "You are not authorized to access this resource")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func BlockIfNotAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if !exists || auth == false {
			ctx.Abort()
			return
		}

		isAdminAny, exists := ctx.Get("isAdmin")
		if !exists || isAdminAny == nil {
			ctx.Abort()
			utils.RenderError(ctx, http.StatusForbidden, "You are not authorized to access this resource")
			return
		}
		isAdmin, ok := isAdminAny.(bool)
		if !ok || !isAdmin {
			ctx.Abort()
			utils.RenderError(ctx, http.StatusForbidden, "You are not authorized to access this resource")
			return
		}

		ctx.Next()
	}
}

func RenderAuthPageIfUnauthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if !exists || auth == false {
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.Index())
			return
		}
		ctx.Next()
	}
}

func RenderMainPageIfAuthenticated(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, exists := ctx.Get("auth")
		if (!exists || auth == true) && ctx.Request.URL.Path == "/auth" {
			ctx.Abort()
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		ctx.Next()
	}
}
