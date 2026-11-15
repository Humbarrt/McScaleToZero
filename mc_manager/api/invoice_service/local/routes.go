package local

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	local_view "mc_manager/web/views/invoice_service/local"
	utils "mc_manager/web/web_utils"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func getLocalConfigHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
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

		payingPlayerId := l.GetPayingPlayerId()
		currencies := l.GetCurrencies()
		currency := l.GetCurrency()

		unsyncedBills, err := queries.GetUnsyncedBills(app.Db)
		if err != nil {
			logging.Error("Error fetching unsynced bills", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		isSynced := len(unsyncedBills) == 0

		hasSaved := ctx.Query("localConfigSuccessIcon") == "true"
		utils.Render(ctx, http.StatusOK, local_view.Config(isAdmin, hasSaved, players, payingPlayerId, currencies, currency, isSynced))
	}
}

func postLocalConfigHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			InvoiceCurrency string `form:"invoiceCurrency"`
			PayingPlayerId  string `form:"payingPlayerId"`
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

		invoiceCurrency := form.InvoiceCurrency
		if invoiceCurrency == "" || !slices.Contains(l.GetCurrencies(), invoiceCurrency) {
			logging.Error("Invalid currency", "currency", invoiceCurrency)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid currency")
			return
		}

		var payingPlayerId *int64
		for _, player := range players {
			if fmt.Sprintf("%d", player.Id) == form.PayingPlayerId {
				payingPlayerId = &player.Id
				break
			}
		}

		if payingPlayerId == nil {
			logging.Error("No paying player configured")
			utils.RenderError(ctx, http.StatusBadRequest, "No paying player configured")
			return
		}

		err = l.UpdateConfig(invoiceCurrency, payingPlayerId)
		if err != nil {
			logging.Error("Error updating config", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		ctx.Redirect(http.StatusSeeOther, ctx.Request.URL.Path+"?localConfigSuccessIcon=true")
	}
}

func getLocalOverviewHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		players, err := queries.GetPlayers(app.Db)
		if err != nil {
			logging.Error("Error fetching players", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		overviewOptionString, _ := ctx.GetQuery("overviewOption")
		var selectedOverviewOption *int64
		if overviewOptionString != "" {
			for _, player := range players {
				if fmt.Sprintf("%d", player.Id) == overviewOptionString {
					selectedOverviewOption = &player.Id
					break
				}
			}
		}

		if selectedOverviewOption == nil {
			renderOverview(ctx, app, l, players)
		} else {
			renderPlayerview(ctx, app, l, players, *selectedOverviewOption)
		}
	}
}

func renderOverview(ctx *gin.Context, app *model.App, l *LocalInvoiceService, players []*model.Player) {
	isAdmin, err := utils.IsAdmin(ctx)
	if err != nil {
		logging.Error("Error checking admin status", "error", err)
		utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}
	currencySymbol := model.LocalInvoiceServiceSupportedCurrencies[l.GetCurrency()]
	payingPlayerId := l.GetPayingPlayerId()

	invoicePlayers, totalSpent, totalPaid, outstandingDebt, invoiceCount, paymentCount := l.GetCache()
	invoicePlayersResponse := []*model.LocalInvoiceServicePlayerCacheResponse{}
	for _, player := range players {
		p := &model.LocalInvoiceServicePlayerCacheResponse{
			Player:         player,
			IsPayingPlayer: payingPlayerId != nil && *payingPlayerId == player.Id,
		}
		if cache, ok := invoicePlayers[player.Id]; ok {
			p.Cache = cache
		} else {
			p.Cache = &model.LocalInvoiceServicePlayerCache{
				TotalSpent:      decimal.Zero,
				TotalPaid:       decimal.Zero,
				OutstandingDebt: decimal.Zero,
				PaymentCount:    0,
			}
		}
		if payingPlayerId != nil && *payingPlayerId == player.Id {
			p.Cache.OutstandingDebt = decimal.Zero
		}

		invoicePlayersResponse = append(invoicePlayersResponse, p)
	}

	cacheResponse := &model.LocalInvoiceServiceCacheResponse{
		Players:         invoicePlayersResponse,
		TotalSpent:      totalSpent,
		TotalPaid:       totalPaid,
		OutstandingDebt: outstandingDebt,
		InvoiceCount:    invoiceCount,
		PaymentCount:    paymentCount,
	}

	hasSaved := ctx.Query("localPaymentSuccessIcon") == "true"
	utils.Render(ctx, http.StatusOK, local_view.Overview(l.SwapEvents, isAdmin, hasSaved, cacheResponse, players, currencySymbol))
}

func renderPlayerview(ctx *gin.Context, app *model.App, l *LocalInvoiceService, players []*model.Player, selectedPlayerId int64) {
	isAdmin, err := utils.IsAdmin(ctx)
	if err != nil {
		logging.Error("Error checking admin status", "error", err)
		utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}
	currencySymbol := model.LocalInvoiceServiceSupportedCurrencies[l.GetCurrency()]
	payingPlayerId := l.GetPayingPlayerId()

	invoicePlayers, _, _, _, _, _ := l.GetCache()
	playerCache := invoicePlayers[selectedPlayerId]
	if playerCache == nil {
		playerCache = &model.LocalInvoiceServicePlayerCache{
			TotalSpent:      decimal.Zero,
			TotalPaid:       decimal.Zero,
			OutstandingDebt: decimal.Zero,
			PaymentCount:    0,
		}
	}
	if payingPlayerId != nil && selectedPlayerId == *payingPlayerId {
		playerCache.OutstandingDebt = decimal.Zero
	}

	history := []*model.LocalInvoiceServicePlayerHistoryResponse{}
	invoices, err := queries.GetLocalInvoiceServiceInvoiceItemsByPlayerId(app.Db, selectedPlayerId)
	if err != nil {
		logging.Error("Error fetching invoices", "error", err)
		utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	payments, err := queries.GetLocalInvoiceServicePaymentsByPlayerId(app.Db, selectedPlayerId)
	if err != nil {
		logging.Error("Error fetching payments", "error", err)
		utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	for _, invoice := range invoices {
		history = append(history, &model.LocalInvoiceServicePlayerHistoryResponse{
			Invoice:     invoice,
			Payment:     nil,
			SortingDate: invoice.CreatedAt,
		})
	}

	for _, payment := range payments {
		sortingDate := payment.CreatedAt
		if payment.EditedAt != nil {
			sortingDate = *payment.EditedAt
		}
		history = append(history, &model.LocalInvoiceServicePlayerHistoryResponse{
			Invoice:     nil,
			Payment:     payment,
			SortingDate: sortingDate,
		})
	}

	slices.SortFunc(history, func(a, b *model.LocalInvoiceServicePlayerHistoryResponse) int {
		return a.SortingDate.Compare(b.SortingDate)
	})

	hasSaved := ctx.Query("localPaymentSuccessIcon") == "true"
	utils.Render(ctx, http.StatusOK, local_view.PlayerView(l.SwapEvents, isAdmin, hasSaved, players, playerCache, history, currencySymbol, fmt.Sprintf("%d", selectedPlayerId)))
}

func getLocalPaymentHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		paymentIdString := ctx.Param("paymentId")
		var paymentId *int64
		if paymentIdString != "" {
			parsedPaymentId, err := strconv.ParseInt(paymentIdString, 10, 64)
			if err != nil {
				logging.Error("Error parsing paymentId", "error", err)
				utils.RenderError(ctx, http.StatusBadRequest, "Invalid paymentId")
				return
			}
			paymentId = &parsedPaymentId
		}
		logging.Info("Fetching local payment", "paymentId", paymentIdString, "parsedPaymentId", paymentId)

		players, err := queries.GetPlayers(app.Db)
		if err != nil {
			logging.Error("Error fetching players", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		preselectedPlayerString, _ := ctx.GetQuery("preselectedPlayer")
		var preselectedPlayerId *int64
		if preselectedPlayerString != "" {
			for _, player := range players {
				if fmt.Sprintf("%d", player.Id) == preselectedPlayerString {
					preselectedPlayerId = &player.Id
					break
				}
			}
		}

		currencySymbol := model.LocalInvoiceServiceSupportedCurrencies[l.GetCurrency()]
		payingPlayerId := l.GetPayingPlayerId()
		if preselectedPlayerId == nil || (payingPlayerId != nil && *preselectedPlayerId == *payingPlayerId) {
			for _, player := range players {
				if payingPlayerId == nil || player.Id != *payingPlayerId {
					preselectedPlayerId = &player.Id
					break
				}
			}
		}

		initialAmount := decimal.Zero
		description := ""
		if paymentId != nil {
			payment, err := queries.GetLocalInvoiceServicePaymentById(app.Db, *paymentId)
			if err != nil {
				logging.Error("Error fetching payment", "error", err)
				utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
				return
			}
			if payingPlayerId != nil && payment.PlayerId == *payingPlayerId {
				logging.Error("Paying player cannot pay themselves", "payingPlayerId", *payingPlayerId)
				utils.RenderError(ctx, http.StatusBadRequest, "Paying player cannot pay themselves")
				return
			}

			logging.Info("Editing local payment", "paymentId", *paymentId, "playerId", payment.PlayerId, "amount", payment.Amount, "description", payment.Description)

			preselectedPlayerId = &payment.PlayerId
			initialAmount = payment.Amount
			description = payment.Description
		}

		invoicePlayers, _, _, _, _, _ := l.GetCache()
		invoicePlayersResponse := []*model.LocalInvoiceServicePlayerCacheResponse{}
		for _, player := range players {
			p := &model.LocalInvoiceServicePlayerCacheResponse{
				Player:         player,
				IsPayingPlayer: payingPlayerId != nil && *payingPlayerId == player.Id,
			}
			if cache, ok := invoicePlayers[player.Id]; ok {
				p.Cache = cache
			} else {
				p.Cache = &model.LocalInvoiceServicePlayerCache{
					TotalSpent:      decimal.Zero,
					TotalPaid:       decimal.Zero,
					OutstandingDebt: decimal.Zero,
					PaymentCount:    0,
				}
			}
			invoicePlayersResponse = append(invoicePlayersResponse, p)
		}

		utils.Render(ctx, http.StatusOK, local_view.Payment(paymentId, invoicePlayersResponse, currencySymbol, *preselectedPlayerId, initialAmount, description))
	}
}

type parsedPaymentForm struct {
	Amount      decimal.Decimal
	PlayerId    int64
	Description string
}

func parsePaymentForm(ctx *gin.Context, app *model.App, l *LocalInvoiceService) (parsedPaymentForm, error) {
	result := parsedPaymentForm{
		Amount:      decimal.Zero,
		PlayerId:    0,
		Description: "",
	}

	var form struct {
		Amount      string `form:"paymentAmount"`
		PlayerId    string `form:"paymentPlayer"`
		Description string `form:"paymentDescription"`
	}
	if err := ctx.ShouldBind(&form); err != nil {
		logging.Error("Error binding form", "error", err)
		utils.RenderError(ctx, http.StatusBadRequest, "Invalid form")
		return result, fmt.Errorf("invalid form")
	}

	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error fetching players", "error", err)
		utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
		return result, fmt.Errorf("error fetching players")
	}

	payingPlayerId := l.GetPayingPlayerId()
	if payingPlayerId != nil && fmt.Sprintf("%d", *payingPlayerId) == form.PlayerId {
		logging.Error("Paying player cannot pay themselves", "payingPlayerId", *payingPlayerId)
		utils.RenderError(ctx, http.StatusBadRequest, "Paying player cannot pay themselves")
		return result, fmt.Errorf("paying player cannot pay themselves")
	}

	var paymentPlayer *model.Player
	for _, player := range players {
		if fmt.Sprintf("%d", player.Id) == form.PlayerId {
			paymentPlayer = player
			break
		}
	}

	if paymentPlayer == nil {
		logging.Error("No paying player configured")
		utils.RenderError(ctx, http.StatusBadRequest, "Player not found")
		return result, fmt.Errorf("player not found")
	}

	// Validate description
	if len(form.Description) > 100 {
		logging.Error("Description too long", "length", len(form.Description))
		utils.RenderError(ctx, http.StatusBadRequest, "Description too long")
		return result, fmt.Errorf("description too long")
	}

	// Validate amount
	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		logging.Error("Error converting amount to decimal", "error", err)
		utils.RenderError(ctx, http.StatusBadRequest, "Invalid amount value")
		return result, fmt.Errorf("invalid amount value")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		logging.Error("Amount must be greater than zero", "amount", amount)
		utils.RenderError(ctx, http.StatusBadRequest, "Amount must be greater than zero")
		return result, fmt.Errorf("amount must be greater than zero")
	} else if amount.GreaterThan(decimal.NewFromInt(1000)) {
		logging.Error("Amount must be less than or equal to 1000", "amount", amount)
		utils.RenderError(ctx, http.StatusBadRequest, "Amount must be less than or equal to 1000")
		return result, fmt.Errorf("amount must be less than or equal to 1000")
	}

	return parsedPaymentForm{
		Amount:      amount,
		PlayerId:    paymentPlayer.Id,
		Description: form.Description,
	}, nil
}

func postLocalPaymentHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		parsedForm, err := parsePaymentForm(ctx, app, l)
		if err != nil {
			return
		}

		err = queries.AddLocalInvoiceServicePayment(app.Db, parsedForm.PlayerId, parsedForm.Amount, parsedForm.Description)
		if err != nil {
			logging.Error("Error adding payment", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		l.RefreshCache()
		app.EB.Broadcast(model.EB_EVENT_INVOICE_CONFIG_UPDATE)
		ctx.Redirect(http.StatusSeeOther, "/is/local/overview?localPaymentSuccessIcon=true")
	}
}

func patchLocalPaymentHandler(app *model.App, l *LocalInvoiceService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		paymentIdString := ctx.Param("paymentId")
		paymentId, err := strconv.ParseInt(paymentIdString, 10, 64)
		if err != nil {
			logging.Error("Error parsing paymentId", "error", err)
			utils.RenderError(ctx, http.StatusBadRequest, "Invalid paymentId")
			return
		}

		payment, err := queries.GetLocalInvoiceServicePaymentById(app.Db, paymentId)
		if err != nil {
			logging.Error("Error fetching payment", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		parsedForm, err := parsePaymentForm(ctx, app, l)
		if err != nil {
			return
		}
		payment.Amount = parsedForm.Amount
		payment.Description = parsedForm.Description

		err = queries.UpdateLocalInvoiceServicePayment(app.Db, payment)
		if err != nil {
			logging.Error("Error adding payment", "error", err)
			utils.RenderError(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}

		l.RefreshCache()
		app.EB.Broadcast(model.EB_EVENT_INVOICE_CONFIG_UPDATE)
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/is/local/overview?overviewOption=%d&localPaymentSuccessIcon=true", payment.PlayerId))
	}
}
