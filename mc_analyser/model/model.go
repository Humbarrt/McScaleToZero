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
	GetAWSProjectName() string
	GetMCServerSSHUser() string
	GetMCServerSSHKey() []byte
	GetDashboardPort() int
	GetTransferServerPort() int
	GetInvoiceService() InvoiceService
	GetVersion() string
	GetDataPath() string

	GetDashboardURL() string
	SetDashboardURL(url string)
	GetMCServerPort() int
	SetMCServerPort(port int) error
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

type Invoice struct {
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	BillId    int       `json:"bill_id"`
	TotalCost int       `json:"total_cost"`
}

type InvoiceServiceUser struct {
	Id    string  `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type CostEstimate struct {
	TotalCost    int    `json:"total_cost"`
	FixedCost    int    `json:"fixed_cost"`
	VariableCost int    `json:"variable_cost"`
	Currency     string `json:"currency"`
	Players      []PlayerCostEstimate
}

type PlayerCostEstimate struct {
	PlayerId int `json:"player_id"`
	Cost     int `json:"cost"`
	Playtime int `json:"playtime"`
	CPUTime  int `json:"cpu_time"`
}

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
	IS         InvoiceServiceAPI
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
