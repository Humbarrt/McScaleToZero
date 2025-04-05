package model

import (
	"embed"
	"mc_analyser/db"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	AWS_ACCESS_KEY_ID      string
	AWS_SECRET_ACCESS_KEY  string
	AWS_INSTANCE_NAME      string
	MC_PORT                int
	SERVER_WATCHER_RETRIES int
}

type AuthClaims struct {
	McUUID string
	jwt.RegisteredClaims
}

type Player struct {
	Id          int64    `json:"id"`
	McUUID      UUID     `json:"mc_uuid"`
	IsAllowed   bool     `json:"is_active"`
	Aliases     []string `json:"aliases"`
	SplitwiseId string   `json:"splitwise_id"`
}

type PlayerEvent struct {
	Id        int64     `json:"id"`
	PlayerId  int64     `json:"player_id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type ServerEvent struct {
	Id        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type App struct {
	Router    *gin.Engine
	FS        embed.FS
	Db        *db.DB
	Config    *Config
	Timeout   time.Duration
	SW        ServerWatcher
	AWS       AWSAPI
	JWTSecret []byte

	//protected
	mutex       sync.Mutex
	_lastUpdate time.Time
}

func (app *App) SetLastUpdate(t time.Time) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app._lastUpdate = t
}
func (app *App) GetLastUpdate() time.Time {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	return app._lastUpdate
}
