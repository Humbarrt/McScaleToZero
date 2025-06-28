package events

import (
	"io"
	"mc_manager/model"
	"mc_manager/web/routes/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(route *gin.Engine, app *model.App) {
	events := route.Group("/events")
	{
		events.GET("/", middleware.BlockIfUnauthenticated(), app.EB.Middleware(), getEventsHandler(app))
	}
}

func getEventsHandler(app *model.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, ok := ctx.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(model.EventClientChan)
		if !ok {
			return
		}

		ctx.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if eventId, ok := <-clientChan; ok {
				ctx.SSEvent(eventId, "")
				return true
			}
			return false
		})
	}
}
