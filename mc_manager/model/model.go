package model

import (
	"embed"
	"mc_manager/db"
	"time"

	"github.com/gin-gonic/gin"
)

type Timeframe struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type App struct {
	Router     *gin.Engine
	FS         embed.FS
	Db         *db.DB
	Config     Config
	WebTimeout time.Duration
	SW         ServerWatcher
	CW         CostWatcher
	AWS        AWSAPI
	IS         InvoiceService
	LW         LogWatcher
	SM         SessionManager
	EB         EventBroadcaster
	JWTSecret  []byte
}
