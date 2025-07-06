package event_broadcaster

import (
	"fmt"
	"log"
	"mc_manager/logging"
	"mc_manager/model"

	"github.com/gin-gonic/gin"
)

type EventBroadcaster struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message chan string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}
}

func (stream *EventBroadcaster) Broadcast(event model.EB_EventType) {
	stream.Message <- fmt.Sprintf("%d", event)
}

func (stream *EventBroadcaster) Start() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				select {
				case clientMessageChan <- eventMsg:
					// Message sent successfully
					/* default:
					logging.Error("EventBroadcaster", "error", "failed to send message to client", "client", clientMessageChan)
					delete(stream.TotalClients, clientMessageChan)
					close(clientMessageChan)
					log.Printf("Client closed. %d registered clients", len(stream.TotalClients)) */
				}
			}
		}
	}
}

func (stream *EventBroadcaster) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(model.EventClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			//<-c.Writer.CloseNotify()

			// Drain client channel so that it does not block. Server may keep sending messages to this channel
			logging.Debug("EventBroadcaster middleware closing client connection")
			_, ok := <-clientChan
			if !ok {
				logging.Debug("EventBroadcaster middleware client channel already closed")
				return
			}

			// Send closed connection to event server
			stream.ClosedClients <- clientChan
			logging.Debug("EventBroadcaster middleware closed client connection")

		}()

		c.Set("clientChan", clientChan)
		c.Next()
	}
}
