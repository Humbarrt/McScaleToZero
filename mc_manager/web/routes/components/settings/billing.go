package settings

import (
	"context"
	"mc_manager/logging"
	"mc_manager/model"
	"net/http"
	"time"

	settings_view "mc_manager/web/views/components/settings"
	utils "mc_manager/web/web_utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func getBillingConfigHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		isAdmin, err := utils.IsAdmin(ctx)
		if err != nil {
			logging.Error("Error checking admin status", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		hasSaved := ctx.Query("billingConfigSuccessIcon") == "true"

		taxRate := app.Config.GetTaxRate()
		networkRatio := app.Config.GetNetworkRatio()
		fallbackFixedCost := app.Config.GetFallbackFixedCost()
		fallbackVariableCost := app.Config.GetFallbackVariableCost()

		utils.Render(ctx, http.StatusOK, settings_view.BillingConfig(isAdmin, taxRate, networkRatio, fallbackFixedCost, fallbackVariableCost, hasSaved))
	}
}

func postBillingConfigHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		var form struct {
			TaxRate              string `form:"taxRate"`
			NetworkRatio         string `form:"networkRatio"`
			FallbackFixedCost    string `form:"fallbackFixedCost"`
			FallbackVariableCost string `form:"fallbackVariableCost"`
		}
		if err := ctx.ShouldBind(&form); err != nil {
			logging.Error("Error binding form", "error", err)
			ctx.String(http.StatusBadRequest, "Invalid form")
			return
		}

		taxRate, err := decimal.NewFromString(form.TaxRate)
		if err != nil {
			logging.Error("Error converting taxRate to decimal", "error", err)
			ctx.String(http.StatusBadRequest, "Invalid taxRate value")
			return
		}

		networkRatio, err := decimal.NewFromString(form.NetworkRatio)
		if err != nil {
			logging.Error("Error converting networkRatio to decimal", "error", err)
			ctx.String(http.StatusBadRequest, "Invalid networkRatio value")
			return
		}

		fallbackFixedCost, err := decimal.NewFromString(form.FallbackFixedCost)
		if err != nil {
			logging.Error("Error converting fallbackFixedCost to decimal", "error", err)
			ctx.String(http.StatusBadRequest, "Invalid fallbackFixedCost value")
			return
		}

		fallbackVariableCost, err := decimal.NewFromString(form.FallbackVariableCost)
		if err != nil {
			logging.Error("Error converting fallbackVariableCost to decimal", "error", err)
			ctx.String(http.StatusBadRequest, "Invalid fallbackVariableCost value")
			return
		}

		taxRate = taxRate.Div(decimal.NewFromFloat(100.0))
		if taxRate.LessThan(decimal.Zero) || taxRate.GreaterThan(decimal.NewFromFloat(1.0)) {
			logging.Error("Invalid taxRate value", "taxRate", taxRate)
			ctx.String(http.StatusBadRequest, "Invalid taxRate value")
			return
		}

		networkRatio = networkRatio.Div(decimal.NewFromFloat(100.0))
		if networkRatio.LessThan(decimal.Zero) || networkRatio.GreaterThan(decimal.NewFromFloat(1.0)) {
			logging.Error("Invalid networkRatio value", "networkRatio", networkRatio)
			ctx.String(http.StatusBadRequest, "Invalid networkRatio value")
			return
		}

		if fallbackFixedCost.LessThan(decimal.Zero) || fallbackFixedCost.GreaterThan(decimal.NewFromFloat(10000.0)) {
			logging.Error("Invalid fallbackFixedCost value", "fallbackFixedCost", fallbackFixedCost)
			ctx.String(http.StatusBadRequest, "Invalid fallbackFixedCost value")
			return
		}

		if fallbackVariableCost.LessThan(decimal.Zero) || fallbackVariableCost.GreaterThan(decimal.NewFromFloat(10000.0)) {
			logging.Error("Invalid fallbackVariableCost value", "fallbackVariableCost", fallbackVariableCost)
			ctx.String(http.StatusBadRequest, "Invalid fallbackVariableCost value")
			return
		}

		err = app.Config.SetTaxRate(taxRate)
		if err != nil {
			logging.Error("Error setting taxRate", "error", err)
			ctx.String(http.StatusInternalServerError, "Error setting taxRate")
			return
		}

		err = app.Config.SetNetworkRatio(networkRatio)
		if err != nil {
			logging.Error("Error setting networkRatio", "error", err)
			ctx.String(http.StatusInternalServerError, "Error setting networkRatio")
			return
		}

		err = app.Config.SetFallbackFixedCost(fallbackFixedCost)
		if err != nil {
			logging.Error("Error setting fallbackFixedCost", "error", err)
			ctx.String(http.StatusInternalServerError, "Error setting fallbackFixedCost")
			return
		}

		err = app.Config.SetFallbackVariableCost(fallbackVariableCost)
		if err != nil {
			logging.Error("Error setting fallbackVariableCost", "error", err)
			ctx.String(http.StatusInternalServerError, "Error setting fallbackVariableCost")
			return
		}

		newTaxRate := app.Config.GetTaxRate()
		newNetworkRatio := app.Config.GetNetworkRatio()
		newFallbackFixedCost := app.Config.GetFallbackFixedCost()
		newFallbackVariableCost := app.Config.GetFallbackVariableCost()

		utils.Render(ctx, http.StatusOK, settings_view.BillingConfig(true, newTaxRate, newNetworkRatio, newFallbackFixedCost, newFallbackVariableCost, true))
		go func() {
			time.Sleep(100 * time.Millisecond)
			app.EB.Broadcast(model.EB_EVENT_BILLING_CONFIG_UPDATE)
		}()
	}
}
