package auth

import (
	"context"
	"database/sql"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mileusna/useragent"
	"golang.org/x/crypto/bcrypt"
)

func signInPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		alias := ctx.Query("alias")
		if alias == "" {
			logging.Error("SignInPageHandler", "error, alias is empty")
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("username is empty")))
			return
		}

		player, err := queries.GetPlayerByAlias(app.Db, alias)
		if err != nil || player == nil {
			logging.Error("SignInPageHandler", "error", err, "player", player)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("player not found")))
			return
		}

		clientKey, err := queries.GetClientKey(app.Db, player.Id)
		if err != nil && err != sql.ErrNoRows {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("internal error, please try again")))
			return
		} else if err == sql.ErrNoRows {
			ctx.Abort()
			session, err := app.SM.CreatePasswordResetSession(player.Id)
			if err != nil || session == nil {
				logging.Error("SignInPageHandler", "error", err)
				utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("internal error, please try again")))
				return
			}

			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(player, session, nil))
			return
		}

		utils.Render(ctx, http.StatusOK, auth_view.SignInForm(player, clientKey, nil))
	}
}

func signInHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		logging.Info("SignInHandler", "request", ctx.Request.URL.String())

		var form struct {
			PlayerId       string `form:"playerId" json:"playerId"`
			HashedPassword string `form:"hashedPassword" json:"hashedPassword"`
		}

		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("invalid request")))
			return
		}

		hashedPassword := []byte(form.HashedPassword)
		if form.PlayerId == "" || len(hashedPassword) == 0 {
			logging.Error("SignInPageHandler", "error", "missing fields", "playerId", form.PlayerId, "hashedPassword", form.HashedPassword)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("invalid request, please reload page")))
			return
		}

		if len(hashedPassword) < 8 || len(hashedPassword) > 256 {
			logging.Error("SignInPageHandler", "error", "invalid password length", "hashedPassword length", len(hashedPassword))
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("invalid request, please reload page")))
			return
		}

		playerId, err := strconv.ParseInt(form.PlayerId, 10, 64)
		if err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("invalid request")))
			return
		}

		player, err := queries.GetPlayerById(app.Db, playerId)
		if err != nil || player == nil {
			logging.Error("SignInPageHandler", "error", err, "player", player)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("internal error, please try again")))
			return
		}

		clientKey, err := queries.GetClientKey(app.Db, player.Id)
		if err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.AliasForm(fmt.Errorf("internal error, please try again")))
			return
		}

		encryptedPassword, err := queries.GetPlayerPassword(app.Db, player.Id)
		if err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.SignInForm(player, clientKey, fmt.Errorf("internal error, please try again")))
			return
		}

		if encryptedPassword == nil {
			logging.Error("SignInPageHandler", "error", "password is nil")
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.SignInForm(player, clientKey, fmt.Errorf("internal error, please try again")))
			return
		}

		err = bcrypt.CompareHashAndPassword(encryptedPassword, hashedPassword)
		if err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.SignInForm(player, clientKey, fmt.Errorf("invalid password")))
			return
		}

		if err := createCookie(ctx, app, player.Id); err != nil {
			logging.Error("SignInPageHandler", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.SignInForm(player, clientKey, fmt.Errorf("internal error, please try again")))
			return
		}

		utils.Render(ctx, http.StatusOK, auth_view.SignInSuccess())
	}
}

func createCookie(ctx *gin.Context, app *model.App, playerId int64) error {
	ua := useragent.Parse(ctx.Request.UserAgent())
	session, err := app.SM.CreateUserSession(playerId, &ua)
	if err != nil || session == nil {
		logging.Error("createCookie", "error", err, "session", session)
		return err
	}

	sessionId := session.GetId()

	claims := model.AuthClaims{
		SessionId: sessionId,
		PlayerId:  playerId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(365 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(app.JWTSecret)
	if err != nil {
		logging.Error("createCookie", "error", err)
		return err
	}

	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie("session", tokenString, 3600*24*365, "/", "", false, true)
	ctx.Status(http.StatusOK)
	return nil
}
