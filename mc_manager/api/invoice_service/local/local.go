package local

import (
	"fmt"
	"maps"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes/middleware"
	local_view "mc_manager/web/views/invoice_service/local"
	"slices"
	"sync"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	exchangerates "github.com/yusufthedragon/exchange-rates-go"
)

type LocalInvoiceService struct {
	app *model.App

	mutex             sync.Mutex
	_currency         string
	_paying_player_id *int64
	_currencies       []string
	_state            model.InvoiceServiceStatus

	//cache
	_invoicePlayers       map[int64]*model.LocalInvoiceServicePlayerCache
	_totalSpent           decimal.Decimal
	_totalPaid            decimal.Decimal
	_totalOutstandingDebt decimal.Decimal
	_paymentCount         int
	_invoiceCount         int

	SwapEvents []model.EB_EventType
}

func NewLocalInvoiceService(app *model.App) *LocalInvoiceService {
	return &LocalInvoiceService{
		app:                   app,
		_state:                model.StatusInvoiceServiceNotInitialized,
		_invoicePlayers:       make(map[int64]*model.LocalInvoiceServicePlayerCache),
		_totalSpent:           decimal.Zero,
		_totalPaid:            decimal.Zero,
		_totalOutstandingDebt: decimal.Zero,

		SwapEvents: []model.EB_EventType{model.EB_EVENT_INVOICE_CONFIG_UPDATE, model.EB_EVENT_NEW_PLAYER, model.EB_EVENT_BILLS_SYNCED},
	}
}

func (l *LocalInvoiceService) Init() error {
	l.mutex.Lock()
	if l._state == model.StatusInvoiceServicePending || l._state == model.StatusInvoiceServiceReady {
		return nil
	}
	l._state = model.StatusInvoiceServicePending

	config, err := queries.GetLocalInvoiceServiceConfig(l.app.Db)
	if err != nil {
		logging.Error("Error getting Local Invoice Service config", "error", err)
		return fmt.Errorf("error getting Local Invoice Service config: %w", err)
	}

	l._currency = config.InvoiceCurrency
	l._paying_player_id = config.PayingPlayerId
	l._currencies = slices.Collect(maps.Keys(model.LocalInvoiceServiceSupportedCurrencies))
	l._state = model.StatusInvoiceServiceReady
	l.mutex.Unlock()

	l.RefreshCache()
	return nil
}

func (l *LocalInvoiceService) Status() model.InvoiceServiceStatus {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l._state
}

func (l *LocalInvoiceService) SyncBills() error {
	if l._paying_player_id == nil {
		return fmt.Errorf("no paying player configured")
	}

	unsyncedBills, err := queries.GetUnsyncedBills(l.app.Db)
	if err != nil {
		logging.Error("Error fetching unsynced bills", "error", err)
		return err
	}

	if len(unsyncedBills) == 0 {
		logging.Debug("No unsynced bills found")
		return nil
	}

	//sort synced bills by billing month ascending
	slices.SortFunc(unsyncedBills, func(a, b *model.Bill) int {
		return a.BillingMonth.EndTime().Compare(b.BillingMonth.EndTime())
	})

	l.mutex.Lock()
	convertToCurrency := l._currency
	l.mutex.Unlock()
	currencyRate := decimal.NewFromInt(1) // Default to 1.0 for USD
	if convertToCurrency != "USD" {
		logging.Debug("Converting currency from USD to", convertToCurrency)
		currencyRateFloat, err := exchangerates.ConvertCurrency(&exchangerates.RequestParameter{
			From:  "USD",
			To:    convertToCurrency,
			Value: 1,
		})

		if err != nil {
			logging.Error("Error getting currency rate", "error", err, "from", "USD", "to", convertToCurrency)
			return fmt.Errorf("error getting currency rate: %w", err)
		}

		if currencyRateFloat == 0 {
			logging.Error("No currency rate found", "from", "USD", "to", convertToCurrency)
			return fmt.Errorf("no currency rate found for USD to %s", convertToCurrency)
		}
		logging.Debug("Currency rate found", "rate", currencyRateFloat, "from", "USD", "to", convertToCurrency)
		currencyRate = decimal.NewFromFloat(currencyRateFloat)
	}

	for _, bill := range unsyncedBills {
		totalAmount := decimal.Zero
		billedPlayers, err := queries.GetBilledPlayersByBillId(l.app.Db, bill.Id)
		if err != nil {
			logging.Error("Error fetching billed players for bill", "billId", bill.Id, "error", err)
			return err
		}

		items := []*model.LocalInvoiceServiceInvoiceItem{}
		for _, billedPlayer := range billedPlayers {
			convertedAmount := billedPlayer.TotalAmount.Mul(currencyRate)
			totalAmount = totalAmount.Add(convertedAmount)

			items = append(items, &model.LocalInvoiceServiceInvoiceItem{
				PlayerId: billedPlayer.PlayerId,
				Amount:   convertedAmount,
			})
		}
		description := bill.BillingMonth.String()
		invoiceId, err := queries.AddLocalInvoiceServiceInvoice(l.app.Db, totalAmount, description, items)
		if err != nil {
			logging.Error("Error adding local invoice", "error", err)
			return err
		}

		err = queries.AddInvoiceServiceInvoiceId(l.app.Db, bill.Id, fmt.Sprintf("%d", invoiceId))
		if err != nil {
			logging.Error("Error adding invoice service invoice ID to bill", "error", err, "billId", bill.Id)
			return fmt.Errorf("error adding invoice service invoice ID to bill: %w", err)
		}
	}
	l.RefreshCache()
	l.app.EB.Broadcast(model.EB_EVENT_BILLS_SYNCED)
	return nil
}

func (l *LocalInvoiceService) GetCurrencies() []string {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l._currencies
}

func (l *LocalInvoiceService) GetCurrency() string {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l._currency == "" {
		return "USD" // Default currency if not set
	}
	return l._currency
}

func (l *LocalInvoiceService) GetPayingPlayerId() *int64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l._paying_player_id
}

func (l *LocalInvoiceService) UpdateConfig(currency string, payingPlayerId *int64) error {
	l.mutex.Lock()
	if currency == "" || !slices.Contains(l._currencies, currency) {
		l.mutex.Unlock()
		return fmt.Errorf("invalid currency: %s", currency)
	}
	if payingPlayerId == nil {
		l.mutex.Unlock()
		return fmt.Errorf("no paying player configured")
	}
	payingPlayer, err := queries.GetPlayerById(l.app.Db, *payingPlayerId)
	if err != nil || payingPlayer == nil {
		l.mutex.Unlock()
		logging.Error("Error fetching paying player", "error", err)
		return fmt.Errorf("error fetching paying player")
	}
	l.mutex.Unlock()

	err = queries.UpdateLocalInvoiceServiceConfig(l.app.Db, &model.LocalInvoiceServiceConfig{
		InvoiceCurrency: currency,
		PayingPlayerId:  payingPlayerId,
	})
	if err != nil {
		logging.Error("Error updating LocalInvoiceService config in database", "error", err)
		return fmt.Errorf("error updating LocalInvoiceService config in database: %w", err)
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l._currency = currency
	l._paying_player_id = payingPlayerId
	l.app.EB.Broadcast(model.EB_EVENT_INVOICE_CONFIG_UPDATE)
	return nil
}

func (l *LocalInvoiceService) RefreshCache() error {
	if l._paying_player_id == nil {
		return fmt.Errorf("no paying player configured")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	invoiceCount, err := queries.GetLocalInvoiceServiceInvoiceCount(l.app.Db)
	playerCache, err := queries.CalculateLocalInvoiceServicePlayerCache(l.app.Db)

	players, err := queries.GetPlayers(l.app.Db)
	if err != nil {
		logging.Error("Error fetching players", "error", err)
		return fmt.Errorf("error fetching players: %w", err)
	}

	totalSpent := decimal.Zero
	totalPaid := decimal.Zero
	totalOutstandingDebt := decimal.Zero
	paymentCount := 0
	for _, player := range players {
		if player.Id == *l._paying_player_id {
			playerCache[player.Id].OutstandingDebt = decimal.Zero
		}

		totalSpent = totalSpent.Add(playerCache[player.Id].TotalSpent)
		totalPaid = totalPaid.Add(playerCache[player.Id].TotalPaid)
		totalOutstandingDebt = totalOutstandingDebt.Add(decimal.Max(decimal.Zero, playerCache[player.Id].OutstandingDebt))
		paymentCount += playerCache[player.Id].PaymentCount

		if player.InvoiceServiceUserId != nil && *player.InvoiceServiceUserId != "" {
			continue
		}
		invoiceServiceUserId := fmt.Sprintf("%d", player.Id)
		player.InvoiceServiceUserId = &invoiceServiceUserId
		err = queries.UpdatePlayerInvoiceServiceUserId(l.app.Db, player.Id, player.InvoiceServiceUserId)
		if err != nil {
			logging.Error("Error updating player invoice service user id", "error", err, "playerId", player.Id)
			return fmt.Errorf("error updating player invoice service user id: %w", err)
		}
	}

	l._invoicePlayers = playerCache
	l._totalSpent = totalSpent
	l._totalPaid = totalPaid
	l._totalOutstandingDebt = totalOutstandingDebt
	l._invoiceCount = invoiceCount
	l._paymentCount = paymentCount

	logging.Debug("LocalInvoiceService cache refreshed")
	return nil
}

func (l *LocalInvoiceService) GetCache() (map[int64]*model.LocalInvoiceServicePlayerCache, decimal.Decimal, decimal.Decimal, decimal.Decimal, int, int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	cacheCopy := make(map[int64]*model.LocalInvoiceServicePlayerCache)
	for k, v := range l._invoicePlayers {
		copy := *v
		cacheCopy[k] = &copy
	}
	return cacheCopy, l._totalSpent, l._totalPaid, l._totalOutstandingDebt, l._invoiceCount, l._paymentCount
}

func (l *LocalInvoiceService) Routes(route *gin.RouterGroup, app *model.App) {
	local := route.Group("/local")
	{
		local.GET("/config", getLocalConfigHandler(app, l))
		local.POST("/config", middleware.BlockIfNotAdmin(), postLocalConfigHandler(app, l))
		local.GET("/overview", getLocalOverviewHandler(app, l))
		local.GET("/payment", middleware.BlockIfNotAdmin(), getLocalPaymentHandler(app, l))
		local.GET("/payment/:paymentId", middleware.BlockIfNotAdmin(), getLocalPaymentHandler(app, l))
		local.POST("/payment", middleware.BlockIfNotAdmin(), postLocalPaymentHandler(app, l))
		local.PATCH("/payment/:paymentId", middleware.BlockIfNotAdmin(), patchLocalPaymentHandler(app, l))
	}
}

func (l *LocalInvoiceService) WebLoader() templ.Component {
	return local_view.Loader(l.SwapEvents)
}
