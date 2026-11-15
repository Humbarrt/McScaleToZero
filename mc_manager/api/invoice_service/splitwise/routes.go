package splitwise

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	splitwise_view "mc_manager/web/views/invoice_service/splitwise"
	utils "mc_manager/web/web_utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func getSplitwiseConfigHandler(app *model.App, s *SplitwiseAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		isAdmin, err := utils.IsAdmin(ctx)
		if err != nil {
			logging.Error("Error checking admin status", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		players, err := queries.GetPlayers(app.Db)
		if err != nil {
			logging.Error("Error fetching players", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		none := "None"
		_splitwiseUsers := s.GetUsers()
		splitwiseUsers := append(
			[]model.SplitwiseUser{
				{
					Id:   "none",
					Name: &none,
				},
			},
			_splitwiseUsers...)

		currencies := s.GetCurrencies()
		currency := s.GetCurrency()

		unsyncedBills, err := queries.GetUnsyncedBills(app.Db)
		if err != nil {
			logging.Error("Error fetching unsynced bills", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		isSynced := len(unsyncedBills) == 0

		hasSaved := ctx.Query("splitwiseConfigSuccessIcon") == "true"
		utils.Render(ctx, http.StatusOK, splitwise_view.Overview(isAdmin, hasSaved, players, splitwiseUsers, currencies, currency, isSynced))
	}
}

func postSplitwiseConfigHandler(app *model.App, s *SplitwiseAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			InvoiceCurrency string `form:"invoiceCurrency"`
		}
		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("Error binding form", "error", err)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid form")
			return
		}

		players, err := queries.GetPlayers(app.Db)
		if err != nil {
			logging.Error("Error fetching players", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		splitwiseUsers := s.GetUsers()
		splitwiseUsersMap := make(map[string]model.SplitwiseUser)
		for _, user := range splitwiseUsers {
			splitwiseUsersMap[user.Id] = user
		}

		newPlayerSplitwiseAssignments := make(map[int64]*string)
		for _, player := range players {
			splitwiseUserId := ctx.PostForm(fmt.Sprintf("splitwiseUser_%d", player.Id))
			if splitwiseUserId == "" {
				logging.Error("Splitwise user ID not found in form", "playerId", player.Id)
			}

			_, exists := splitwiseUsersMap[splitwiseUserId]
			if (!exists && player.InvoiceServiceUserId == nil) || (player.InvoiceServiceUserId != nil && *player.InvoiceServiceUserId == splitwiseUserId) {
				continue
			} else if !exists {
				newPlayerSplitwiseAssignments[player.Id] = nil
				continue
			}

			newPlayerSplitwiseAssignments[player.Id] = &splitwiseUserId
		}

		for playerId, splitwiseUserId := range newPlayerSplitwiseAssignments {
			logging.Debug("Updating player invoice service user ID", "playerId", playerId, "splitwiseUserId", splitwiseUserId)
			err := queries.UpdatePlayerInvoiceServiceUserId(app.Db, playerId, splitwiseUserId)
			if err != nil {
				logging.Error("Error updating player invoice service user ID", "playerId", playerId, "splitwiseUserId", splitwiseUserId, "error", err)
				utils.RenderError(ctx, http.StatusInternalServerError, "Error updating player invoice service user ID")
				return
			}
		}

		err = s.SetCurrency(form.InvoiceCurrency)
		if err != nil {
			logging.Error("Error setting splitwise currency", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Error setting splitwise currency")
			return
		}

		ctx.Redirect(http.StatusSeeOther, ctx.Request.URL.Path+"?splitwiseConfigSuccessIcon=true")
		go func() {
			time.Sleep(100 * time.Millisecond)
			app.EB.Broadcast(model.EB_EVENT_INVOICE_CONFIG_UPDATE)
		}()
	}
}

func getSplitwiseUsersHandler(app *model.App, s *SplitwiseAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		err := s.UpdateUsers()
		if err != nil {
			logging.Error("Error updating splitwise users", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		getSplitwiseConfigHandler(app, s)(ctx)
		go func() {
			time.Sleep(100 * time.Millisecond)
			app.EB.Broadcast(model.EB_EVENT_INVOICE_CONFIG_UPDATE)
		}()
	}
}
