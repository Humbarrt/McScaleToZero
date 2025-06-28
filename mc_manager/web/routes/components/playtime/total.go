package playtime

import (
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	playtime_view "mc_manager/web/views/components/playtime"
	utils "mc_manager/web/web_utils"
	"time"

	"github.com/gin-gonic/gin"
)

func computeTotalChartResponse(ctx *gin.Context, app *model.App, players []*model.Player, billedPlayerTime []*model.BilledPlayerTime, currentEstimate *[]model.PlayerCostEstimate) ([]model.TotalPlayerTimeResponse, time.Duration) {
	var playerMap = make(map[int64]struct {
		Playtime time.Duration `json:"playtime"`
		CPUTime  time.Duration `json:"cpu_time"`
	})

	for _, player := range billedPlayerTime {
		if _, exists := playerMap[player.PlayerId]; !exists {
			playerMap[player.PlayerId] = struct {
				Playtime time.Duration `json:"playtime"`
				CPUTime  time.Duration `json:"cpu_time"`
			}{
				Playtime: player.Playtime,
				CPUTime:  player.CPUTime,
			}
		} else {
			playerMap[player.PlayerId] = struct {
				Playtime time.Duration `json:"playtime"`
				CPUTime  time.Duration `json:"cpu_time"`
			}{
				Playtime: playerMap[player.PlayerId].Playtime + player.Playtime,
				CPUTime:  playerMap[player.PlayerId].CPUTime + player.CPUTime,
			}
		}
	}

	if currentEstimate != nil {
		for _, player := range *currentEstimate {
			if _, exists := playerMap[player.PlayerId]; !exists {
				playerMap[player.PlayerId] = struct {
					Playtime time.Duration `json:"playtime"`
					CPUTime  time.Duration `json:"cpu_time"`
				}{
					Playtime: player.Playtime,
					CPUTime:  player.CPUTime,
				}
			} else {
				playerMap[player.PlayerId] = struct {
					Playtime time.Duration `json:"playtime"`
					CPUTime  time.Duration `json:"cpu_time"`
				}{
					Playtime: playerMap[player.PlayerId].Playtime + player.Playtime,
					CPUTime:  playerMap[player.PlayerId].CPUTime + player.CPUTime,
				}
			}
		}
	}

	response := []model.TotalPlayerTimeResponse{}
	maxDuration := time.Duration(0)
	for _, player := range players {
		playerTime := time.Duration(0)
		cpuTime := time.Duration(0)

		if data, exists := playerMap[player.Id]; exists {
			playerTime = data.Playtime
			cpuTime = data.CPUTime
		}

		if playerTime > maxDuration {
			maxDuration = playerTime
		}

		response = append(response, model.TotalPlayerTimeResponse{
			PlayerId: player.Id,
			Aliases:  player.Aliases,
			Playtime: playerTime,
			CPUTime:  cpuTime,
		})
	}

	return response, maxDuration
}

func renderTotalPlaytime(ctx *gin.Context, app *model.App, chartIdString string) {
	billedPlayerTime, err := queries.GetBilledPlayerTime(app.Db)
	if err != nil {
		logging.Error("Error getting billed player time", "error", err)
		ctx.Status(500)
		return
	}

	currentEstimate, err := app.CW.CurrentCostEstimate()
	if err != nil {
		logging.Error("Error getting current cost estimate", "error", err)
		ctx.Status(500)
		return
	}

	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		ctx.Status(500)
		return
	}

	response, maxDuration := computeTotalChartResponse(ctx, app, players, billedPlayerTime, &currentEstimate.Players)
	utils.Render(ctx, 200, playtime_view.TotalChart(chartIdString, true, "Total", response, maxDuration, false, nil))
}
