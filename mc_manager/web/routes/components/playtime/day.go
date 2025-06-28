package playtime

import (
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	playtime_view "mc_manager/web/views/components/playtime"
	utils "mc_manager/web/web_utils"
	"time"

	"github.com/gin-gonic/gin"
)

func renderDayPlaytime(ctx *gin.Context, app *model.App, year int, month int, day int, timezone *time.Location, usingLocalTime bool) {
	isLive := false
	if fmt.Sprintf("%d-%02d-%02d", year, month, day) == time.Now().In(timezone).Format("2006-01-02") {
		isLive = true
	}

	tf := model.Timeframe{
		Start: time.Date(year, time.Month(month), day, 0, 0, 0, 0, timezone),
		End:   time.Date(year, time.Month(month), day, 23, 59, 59, 999999999, timezone),
	}

	events, err := queries.GetPlayersEvents(app.Db, tf)
	if err != nil {
		logging.Error("Error getting player events", "error", err)
		ctx.Status(500)
		return
	}

	players, err := queries.GetPlayers(app.Db)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		ctx.Status(500)
		return
	}

	playersMap := make(map[int64]*model.Player)
	response := map[int64]*model.PlayerTimelineResponse{}
	for _, player := range players {
		response[player.Id] = &model.PlayerTimelineResponse{
			PlayerId: player.Id,
			Aliases:  player.Aliases,
			Events:   []*model.PlayerEvent{},
		}
		playersMap[player.Id] = player
	}

	for _, event := range events {
		event.Timestamp = event.Timestamp.In(timezone)
		if _, ok := response[event.PlayerId]; !ok {
			logging.Warn("Player event for unknown player", "playerId", event.PlayerId)
			continue
		}

		playerTimeline := response[event.PlayerId]
		if len(playerTimeline.Events) == 0 && !event.IsOnline {
			playerTimeline.Events = append(playerTimeline.Events, &model.PlayerEvent{
				Id:        0,
				PlayerId:  event.PlayerId,
				Timestamp: tf.Start,
				IsOnline:  true,
			})
		}

		playerTimeline.Events = append(playerTimeline.Events, event)
		response[event.PlayerId] = playerTimeline
	}

	orderedResponse := make([]*model.PlayerTimelineResponse, 0, len(response))
	for _, player := range players {
		if playerTimeline, exists := response[player.Id]; exists {
			orderedResponse = append(orderedResponse, playerTimeline)
		} else {
			orderedResponse = append(orderedResponse, &model.PlayerTimelineResponse{
				PlayerId: player.Id,
				Aliases:  player.Aliases,
				Events:   []*model.PlayerEvent{},
			})
		}
	}

	eventList := model.EventListResponse{
		Players: playersMap,
		Events:  events,
	}

	utils.Render(ctx, 200, playtime_view.TimelineChart(fmt.Sprintf("%d%02d%02d", year, month, day), isLive, fmt.Sprintf("%d %s %d", year, model.MonthsLabels[month-1], day), orderedResponse, tf, &eventList, usingLocalTime))
}
