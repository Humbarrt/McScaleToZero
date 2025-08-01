package model

import "github.com/gin-gonic/gin"

type EB_EventType int

const (
	EB_EVENT_SERVER_ONLINE EB_EventType = iota
	EB_EVENT_SERVER_READY
	EB_EVENT_SERVER_OFFLINE
	EB_EVENT_NEW_BILL
	EB_EVENT_NEW_COST_ESTIMATE
	EB_EVENT_PLAYER_JOINED
	EB_EVENT_PLAYER_LEFT
	EB_EVENT_SERVER_CONFIG_UPDATE
	EB_EVENT_SERVER_INFO_UPDATE
	EB_EVENT_BILLING_CONFIG_UPDATE
	EB_EVENT_INVOICE_CONFIG_UPDATE
)

type EventBroadcaster interface {
	Broadcast(event EB_EventType)
	Middleware() gin.HandlerFunc
}

type EventClientChan chan string
