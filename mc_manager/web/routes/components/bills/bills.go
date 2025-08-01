package bills

import (
	"context"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	bills_view "mc_manager/web/views/components/bills"
	utils "mc_manager/web/web_utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.RouterGroup, app *model.App) {
	bills := route.Group("/bills")
	{
		bills.GET("/selector", getYearSelector(app))
		bills.GET("/selector/:year", getMonthSelector(app))
		bills.GET("/:month", getBill(app))
	}
}

func getBill(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), app.WebTimeout)
		defer cancel()

		logging.Info("Retrieving bill for month", "month", ctx.Param("month"))
		monthString := ctx.Param("month")
		isPlayerModeString := ctx.Query("isPlayerMode")

		isPlayerMode := true
		if strings.ToLower(isPlayerModeString) == "false" {
			isPlayerMode = false
		}

		isLatest := strings.ToLower(monthString) == "latest"
		var billResponse model.BillResponse
		billPlayersResponse := []model.BillPlayerResponse{}
		billItemsResponse := []model.BillItem{}
		if isLatest {
			costEstimate, err := app.CW.CurrentCostEstimate()
			if err != nil {
				logging.Error("Error retrieving current cost estimate", "error", err)
				ctx.JSON(500, gin.H{"error": "Error retrieving current cost estimate"})
				return
			}

			billResponse = model.BillResponse{
				Month:          model.NewYearMonthFromTime(time.Now()),
				TotalAmount:    costEstimate.TotalCost,
				FixedAmount:    costEstimate.FixedCost,
				VariableAmount: costEstimate.VariableCost,
				ServerUptime:   costEstimate.ServerUptime,
			}

			isPlayerMode = true
			for _, player := range costEstimate.Players {
				billPlayersResponse = append(billPlayersResponse,
					model.BillPlayerResponse{
						PlayerId:       player.PlayerId,
						TotalAmount:    player.TotalCost,
						FixedAmount:    player.FixedCost,
						VariableAmount: player.VariableCost,
						Playtime:       player.Playtime,
						CPUTime:        player.CPUTime,
					},
				)
			}

		} else {
			month, err := model.NewYearMonthFromString(monthString)
			if err != nil {
				logging.Error("Invalid month format", "error", err, "month", monthString)
				ctx.JSON(400, gin.H{"error": "Invalid month format"})
				return
			}

			bill, err := queries.GetFullBillByBillingMonth(app.Db, month)
			if err != nil || bill == nil {
				logging.Error("Error retrieving bill", "error", err, "month", month)
				ctx.JSON(500, gin.H{"error": "Error retrieving bill"})
				return
			}

			billResponse = model.BillResponse{
				Month:          bill.Bill.BillingMonth,
				TotalAmount:    bill.Bill.TotalAmount,
				FixedAmount:    bill.Bill.FixedAmount,
				VariableAmount: bill.Bill.VariableAmount,
				ServerUptime:   bill.Bill.ServerUptime,
			}

			if isPlayerMode {
				for _, player := range bill.BilledPlayers {
					billPlayersResponse = append(billPlayersResponse,
						model.BillPlayerResponse{
							PlayerId:       player.PlayerId,
							TotalAmount:    player.TotalAmount,
							FixedAmount:    player.FixedAmount,
							VariableAmount: player.VariableAmount,
							Playtime:       player.Playtime,
							CPUTime:        player.CPUTime,
						},
					)
				}
			} else {
				for _, item := range bill.BillItems {
					billItemsResponse = append(billItemsResponse,
						model.BillItem{
							Id:      item.Id,
							BillId:  item.BillId,
							Label:   item.Label,
							IsFixed: item.IsFixed,
							Amount:  item.Amount,
						},
					)
				}

				utils.Render(ctx, http.StatusOK, bills_view.BillItems(billResponse, billItemsResponse, isLatest))
				return
			}
		}

		players, err := queries.GetPlayers(app.Db)
		if err != nil {
			logging.Error("Error retrieving players", "error", err)
			ctx.JSON(500, gin.H{"error": "Error retrieving players"})
			return
		}

		playerMap := make(map[int64]*model.Player)
		for _, player := range players {
			playerMap[player.Id] = player
		}

		for i, player := range billPlayersResponse {
			if p, exists := playerMap[player.PlayerId]; exists {
				billPlayersResponse[i].Aliases = p.Aliases
			} else {
				logging.Error("Player not found for bill player", "playerId", player.PlayerId)
			}
		}
		utils.Render(ctx, http.StatusOK, bills_view.BillPlayers(billResponse, billPlayersResponse, isLatest))
	}
}
