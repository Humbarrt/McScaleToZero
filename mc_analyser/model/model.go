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
	FLY_API_BASE_URL string
	FLY_API_TOKEN    string
	FLY_APP_NAME     string
	SIGNING_KEY      []byte
	MC_PORT          int
}

type AuthClaims struct {
	McId string
	jwt.RegisteredClaims
}

type Player struct {
	Id          int64    `json:"id"`
	McId        string   `json:"mc_id"`
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
	Router         *gin.Engine
	FS             embed.FS
	Db             *db.DB
	Config         *Config
	Timeout        time.Duration
	lastUpdate     time.Time
	mu             sync.Mutex
	isServerOnline bool
}

func (app *App) SetServerOnline(online bool) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.isServerOnline = online
}
func (app *App) IsServerOnline() bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.isServerOnline
}
func (app *App) SetLastUpdate(t time.Time) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.lastUpdate = t
}
func (app *App) GetLastUpdate() time.Time {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.lastUpdate
}
