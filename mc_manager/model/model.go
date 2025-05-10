package model

import (
	"embed"
	"mc_manager/db"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
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
	GetTaxRate() decimal.Decimal
	SetTaxRate(taxRate decimal.Decimal) error
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
	TotalCost    decimal.Decimal `json:"total_cost"`
	FixedCost    decimal.Decimal `json:"fixed_cost"`
	VariableCost decimal.Decimal `json:"variable_cost"`
	Currency     string          `json:"currency"`
	ServerUptime time.Duration   `json:"server_uptime"`
	Players      []PlayerCostEstimate
}

type PlayerCostEstimate struct {
	PlayerId int64           `json:"player_id"`
	Cost     decimal.Decimal `json:"cost"`
	Playtime time.Duration   `json:"playtime"`
	CPUTime  time.Duration   `json:"cpu_time"`
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
	LW         LogWatcher
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
