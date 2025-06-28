package auth

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	auth_view "mc_manager/web/views/auth"
	utils "mc_manager/web/web_utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func resetPasswordHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			PlayerId       string `form:"playerId" json:"playerId"`
			SessionId      string `form:"sessionId" json:"sessionId"`
			HashedPassword string `form:"hashedPassword" json:"hashedPassword"`
		}

		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("ResetPassword", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		hashedPassword := []byte(form.HashedPassword)
		if form.PlayerId == "" || form.SessionId == "" || len(hashedPassword) == 0 {
			logging.Error("ResetPassword", "error", "missing fields", "playerId", form.PlayerId, "sessionId", form.SessionId)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		if len(hashedPassword) < 8 || len(hashedPassword) > 256 {
			logging.Error("ResetPassword", "error", "invalid password length", "hashedPassword length", len(hashedPassword))
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		playerId, err := strconv.ParseInt(form.PlayerId, 10, 64)
		if err != nil {
			logging.Error("ResetPassword", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		player, err := queries.GetPlayerById(app.Db, playerId)
		if err != nil || player == nil {
			logging.Error("ResetPassword", "error", err, "player", player)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		encryptedPassword, err := bcrypt.GenerateFromPassword(hashedPassword, bcrypt.DefaultCost)
		if err != nil {
			logging.Error("ResetPassword", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(player, nil, fmt.Errorf("internal error, please try again")))
			return
		}

		if err := app.SM.SetResetPassword(player.Id, form.SessionId, encryptedPassword); err != nil {
			logging.Error("ResetPassword", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(player, nil, fmt.Errorf("internal error, please try again")))
			return
		}

		session, err := app.SM.GetPasswordResetSessionById(player.Id, form.SessionId)
		if err != nil || session == nil {
			logging.Error("ResetPassword", "error", err, "session", session)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.PasswordResetForm(player, nil, fmt.Errorf("session expired, please reload page")))
			return
		}

		utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(player, session, nil))
	}
}

func getResetPasswordPageHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			PlayerId  string `form:"playerId" json:"playerId"`
			SessionId string `form:"sessionId" json:"sessionId"`
		}

		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("GetResetPasswordPage", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		if form.PlayerId == "" || form.SessionId == "" {
			logging.Error("GetResetPasswordPage", "error", "missing fields", "playerId", form.PlayerId, "sessionId", form.SessionId)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		playerId, err := strconv.ParseInt(form.PlayerId, 10, 64)
		if err != nil {
			logging.Error("GetResetPasswordPage", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		player, err := queries.GetPlayerById(app.Db, playerId)
		if err != nil || player == nil {
			logging.Error("GetResetPasswordPage", "error", err, "player", player)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(nil, nil, fmt.Errorf("invalid request, please reload page")))
			return
		}

		session, err := app.SM.GetPasswordResetSessionById(player.Id, form.SessionId)
		if err != nil || session == nil {
			logging.Error("GetResetPasswordPage", "error", err, "session", session)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(player, nil, fmt.Errorf("session expired, please reload page")))
			return
		}

		if !session.GetIsAuthenticated() {
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(player, session, nil))
			return
		}

		session, err = app.SM.ConsumePasswordResetSession(player.Id, form.SessionId)
		if err != nil || session == nil {
			logging.Error("GetResetPasswordPage", "error", err, "session", session)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(player, nil, fmt.Errorf("session expired, please reload page")))
			return
		}

		err = createCookie(ctx, app, player.Id)
		if err != nil {
			logging.Error("GetResetPasswordPage", "error", err)
			ctx.Abort()
			utils.Render(ctx, http.StatusOK, auth_view.WaitForChatToken(player, nil, fmt.Errorf("session expired, please reload page")))
			return
		}

		utils.Render(ctx, http.StatusOK, auth_view.SignInSuccess())
	}
}
