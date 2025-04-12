package model

import (
	"embed"
	"mc_analyser/db"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Config interface {
	GetAWSAccessKeyId() string
	GetAWSSecretAccessKey() string
	GetAWSInstanceName() string
	GetSplitwiseApiKey() string
	GetSplitwiseGroupId() string
	GetMCServerSSHUser() string
	GetMCServerSSHKey() []byte
	GetVersion() string
	GetDataPath() string

	GetMCPort() int
	SetMCPort(port int) error
	GetMCOfflineImage() *[]byte
	SetMCOfflineImage(image *[]byte) error
	GetMCOfflineMOTD() *string
	SetMCOfflineMOTD(motd *string) error
	GetIPCheck() bool
	SetIPCheck(ipCheck bool) error
	GetBackupsToKeep() int
	SetBackupsToKeep(backupsToKeep int) error
}

type AuthClaims struct {
	McUUID string
	jwt.RegisteredClaims
}

type Player struct {
	Id          int      `json:"id"`
	McUUID      UUID     `json:"mc_uuid"`
	IsAllowed   bool     `json:"is_active"`
	IsAdmin     bool     `json:"is_admin"`
	LastIPV4    *string  `json:"last_ip"`
	Aliases     []string `json:"aliases"`
	SplitwiseId *string  `json:"splitwise_id"`
}

type PlayerEvent struct {
	Id        int       `json:"id"`
	PlayerId  int       `json:"player_id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type ServerEvent struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	IsOnline  bool      `json:"is_online"`
}

type Playtime struct {
	PlayerId int           `json:"player_id"`
	Playtime time.Duration `json:"playtime"`
}

type App struct {
	Router     *gin.Engine
	FS         embed.FS
	Db         *db.DB
	Config     Config
	WebTimeout time.Duration
	SW         ServerWatcher
	AWS        AWSAPI
	JWTSecret  []byte

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
