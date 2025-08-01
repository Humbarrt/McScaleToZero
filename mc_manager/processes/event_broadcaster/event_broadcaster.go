package event_broadcaster

import (
	"fmt"
	"mc_manager/logging"
	"mc_manager/model"

	"github.com/gin-gonic/gin"
	sse "github.com/r3labs/sse/v2"
)

type EventBroadcaster struct {
	server *sse.Server
	stream *sse.Stream
}

func NewEventBroadcaster() *EventBroadcaster {
	server := sse.New()
	return &EventBroadcaster{
		server: server,
		stream: server.CreateStream("messages"),
	}
}

func (eb *EventBroadcaster) Broadcast(event model.EB_EventType) {
	logging.Debug("Broadcasting", "event", event)
	eb.server.Publish(eb.stream.ID, &sse.Event{
		Event: fmt.Appendf(nil, "%d", event),
		Data:  fmt.Appendf(nil, "%d", event),
	})
}

func (eb *EventBroadcaster) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Set("stream", eb.stream.ID)
		c.Request.URL.RawQuery = q.Encode()

		eb.server.ServeHTTP(c.Writer, c.Request)
		logging.Debug("The client has disconnected from the SSE stream")
	}
}
