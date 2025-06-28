package splitwise

import (
	"context"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"mc_manager/web/routes/middleware"
	splitwise_view "mc_manager/web/views/invoice_service/splitwise"

	"github.com/a-h/templ"
	spw "github.com/aanzolaavila/splitwise.go"
	"github.com/aanzolaavila/splitwise.go/resources"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	exchangerates "github.com/yusufthedragon/exchange-rates-go"
)

type SplitwiseAPI struct {
	app     *model.App
	token   string
	groupId int

	//protected
	mutex                   sync.Mutex
	_users                  []*model.SplitwiseUser
	_lastUsersDownload      time.Time
	_currency               string
	_currencies             []string
	_lastCurrenciesDownload time.Time
	_client                 *spw.Client
	_state                  model.InvoiceServiceStatus
}

func NewSplitwiseAPI(app *model.App) (*SplitwiseAPI, error) {
	splitwiseToken := os.Getenv("SPLITWISE_API_TOKEN")
	if splitwiseToken == "" {
		return nil, fmt.Errorf("SPLITWISE_API_TOKEN not set")
	}

	splitwiseGroupIdString := os.Getenv("SPLITWISE_GROUP_ID")
	if splitwiseGroupIdString == "" {
		return nil, fmt.Errorf("SPLITWISE_GROUP_ID not set")
	}
	splitwiseGroupId, err := strconv.Atoi(splitwiseGroupIdString)
	if err != nil {
		return nil, fmt.Errorf("SPLITWISE_GROUP_ID is not a number: %s", splitwiseGroupIdString)
	}

	return &SplitwiseAPI{
		app:                     app,
		token:                   splitwiseToken,
		groupId:                 splitwiseGroupId,
		_users:                  []*model.SplitwiseUser{},
		_lastUsersDownload:      time.Time{},
		_currency:               "USD", // Default currency
		_currencies:             []string{},
		_lastCurrenciesDownload: time.Time{},
		_client:                 nil,
		_state:                  model.StatusInvoiceServiceNotInitialized,
	}, nil
}

func (s *SplitwiseAPI) updateState(state model.InvoiceServiceStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s._state = state
}

func (s *SplitwiseAPI) Init() error {
	s.mutex.Lock()
	if s._state == model.StatusInvoiceServicePending || s._state == model.StatusInvoiceServiceReady {
		s.mutex.Unlock()
		return nil
	}

	if s._client != nil {
		s._client.HttpClient = nil
		s._client = nil
	}
	s._state = model.StatusInvoiceServicePending
	s.mutex.Unlock()
	s.updateState(model.StatusInvoiceServicePending)

	s._client = &spw.Client{
		Token: s.token,
	}

	config, err := queries.GetSplitwiseConfig(s.app.Db)
	if err != nil {
		logging.Error("Error getting Splitwise config", "error", err)
		return fmt.Errorf("error getting Splitwise config: %w", err)
	}

	s._currency = config.InvoiceCurrency
	s._currencies = config.Currencies
	s._users = config.Users

	s.updateState(model.StatusInvoiceServiceReady)
	return nil
}

func (s *SplitwiseAPI) Status() model.InvoiceServiceStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s._state
}

func (s *SplitwiseAPI) SyncBills() error {
	unsyncedBills, err := queries.GetUnsyncedBills(s.app.Db)
	if err != nil {
		logging.Error("Error fetching unsynced bills", "error", err)
		return err
	}

	for _, bill := range unsyncedBills {
		if bill == nil {
			continue
		}

		err := s.createInvoice(bill)
		if err != nil {
			logging.Error("Error creating Splitwise invoice", "error", err, "bill", bill)
			return fmt.Errorf("error creating Splitwise invoice: %w", err)
		}
	}

	return nil
}

func (s *SplitwiseAPI) createInvoice(bill *model.Bill) error {
	billedPlayers, err := queries.GetBilledPlayersByBillId(s.app.Db, bill.Id)
	if err != nil {
		logging.Error("Error getting billed players for bill", "error", err, "bill", bill)
		return fmt.Errorf("error getting billed players for bill: %w", err)
	}

	if len(billedPlayers) == 0 {
		logging.Warn("No billed players found for bill", "bill", bill)
		return nil
	}

	players, err := queries.GetPlayers(s.app.Db)
	if err != nil {
		logging.Error("Error getting players", "error", err)
		return fmt.Errorf("error getting players: %w", err)
	}

	playersMap := make(map[int64]*model.Player)
	for _, player := range players {
		if player == nil {
			continue
		}
		playersMap[player.Id] = player
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s._client == nil {
		return fmt.Errorf("splitwise client not initialized")
	}

	currencyRate := decimal.NewFromInt(1) // Default to 1.0 for USD
	if s._currency != "USD" {
		logging.Debug("Converting currency from USD to", s._currency)
		currencyRateFloat, err := exchangerates.ConvertCurrency(&exchangerates.RequestParameter{
			From:  "USD",
			To:    s._currency,
			Value: 1,
		})

		if err != nil {
			logging.Error("Error getting currency rate", "error", err, "from", "USD", "to", s._currency)
			return fmt.Errorf("error getting currency rate: %w", err)
		}

		if currencyRateFloat == 0 {
			logging.Error("No currency rate found", "from", "USD", "to", s._currency)
			return fmt.Errorf("no currency rate found for USD to %s", s._currency)
		}
		logging.Debug("Currency rate found", "rate", currencyRateFloat, "from", "USD", "to", s._currency)
		currencyRate = decimal.NewFromFloat(currencyRateFloat)
	}

	description := fmt.Sprintf("MC Manager Invoice for %s", bill.BillingMonth.String())
	expenseParams := spw.CreateExpenseParams{
		spw.CreateExpenseCurrencyCode: s._currency,
	}

	currentUser, err := s._client.GetCurrentUser(context.Background())
	if err != nil {
		logging.Error("Error getting current Splitwise user", "error", err)
		return fmt.Errorf("error getting current Splitwise user: %w", err)
	}

	expenseUsersMap := map[string]*spw.ExpenseUser{}
	for _, billedPlayer := range billedPlayers {
		if billedPlayer == nil {
			continue
		}
		player, exists := playersMap[billedPlayer.PlayerId]
		if !exists {
			logging.Error("Player not found for billed player", "billedPlayer", billedPlayer)
			return fmt.Errorf("player not found for billed player: %d", billedPlayer.PlayerId)
		}

		if player == nil || player.Id == 0 {
			return fmt.Errorf("invalid player for billed player. Player Id: %d", billedPlayer)
		}

		if player.InvoiceServiceUserId == nil || *player.InvoiceServiceUserId == "" {
			return fmt.Errorf("player %d does not have a Splitwise user ID", player.Id)
		}
		invoiceServiceUserIdInt, err := strconv.ParseUint(*player.InvoiceServiceUserId, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid Splitwise user ID for player %d: %s", player.Id, *player.InvoiceServiceUserId)
		}
		invoiceServiceUserId := resources.UserID(invoiceServiceUserIdInt)
		convertedAmount := billedPlayer.TotalAmount.Mul(currencyRate).RoundCeil(2).InexactFloat64()

		_, exists = expenseUsersMap[*player.InvoiceServiceUserId]
		if exists {
			expenseUsersMap[*player.InvoiceServiceUserId].OwedShare += convertedAmount
			if currentUser.ID == invoiceServiceUserId {
				expenseUsersMap[*player.InvoiceServiceUserId].PaidShare = expenseUsersMap[*player.InvoiceServiceUserId].OwedShare
			}
			continue
		}

		paidShare := 0.0
		if currentUser.ID == invoiceServiceUserId {
			paidShare = convertedAmount
		}
		expenseUsersMap[*player.InvoiceServiceUserId] = &spw.ExpenseUser{
			Id:        invoiceServiceUserId,
			PaidShare: paidShare,
			OwedShare: convertedAmount,
		}
	}

	totalExpense := 0.0
	expenseUsers := []spw.ExpenseUser{}
	for _, user := range expenseUsersMap {
		if user == nil {
			continue
		}
		totalExpense += user.OwedShare
		expenseUsers = append(expenseUsers, *user)
	}

	invoices, err := s._client.CreateExpenseByShares(context.Background(), totalExpense, description, s.groupId, expenseParams, expenseUsers)
	if err != nil {
		logging.Error("Error creating Splitwise expense", "error", err, "bill", bill)
		return fmt.Errorf("error creating Splitwise expense: %w", err)
	}

	if len(invoices) == 0 {
		logging.Error("No invoice created for bill", "bill", bill.BillingMonth.String())
		return fmt.Errorf("no invoice created for bill: %s", bill.BillingMonth.String())
	}

	logging.Info("Splitwise invoice created", "invoice", invoices)
	invoiceId := fmt.Sprintf("%d", invoices[0].ID)
	err = queries.AddInvoiceServiceInvoiceId(s.app.Db, bill.Id, invoiceId)
	if err != nil {
		logging.Error("Error adding invoice service invoice ID to bill", "error", err, "billId", bill.Id)
		return fmt.Errorf("error adding invoice service invoice ID to bill: %w", err)
	}

	logging.Info("Splitwise invoice created successfully", "invoiceId", invoiceId, "billId", bill.Id)
	return nil
}

func (s *SplitwiseAPI) UpdateUsers() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s._lastUsersDownload = time.Now()
	group, err := s._client.GetGroup(context.Background(), s.groupId)
	if err != nil {
		return fmt.Errorf("error getting splitwise group: %s", err)
	}

	var users []*model.SplitwiseUser
	for _, member := range group.Members {
		combinedNameValue := fmt.Sprintf("%s %s", member.FirstName, member.LastName)
		combinedName := &combinedNameValue
		if *combinedName == "" {
			combinedName = nil
		}

		users = append(users, &model.SplitwiseUser{
			Id:    fmt.Sprint(member.ID),
			Name:  combinedName,
			Email: &member.Email,
		})
	}

	err = queries.UpdateSplitwiseConfigUsers(s.app.Db, users)
	if err != nil {
		logging.Error("Error updating Splitwise users in database", "error", err)
		return fmt.Errorf("error updating Splitwise users in database: %w", err)
	}

	s._users = users
	return nil
}

func (s *SplitwiseAPI) GetUsers() []model.SplitwiseUser {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s._users) == 0 && (time.Since(s._lastUsersDownload) > 1*time.Minute || s._lastUsersDownload.IsZero()) {
		s.mutex.Unlock()
		err := s.UpdateUsers()
		s.mutex.Lock()
		if err != nil {
			logging.Error("Error updating Splitwise users", "error", err)
			return []model.SplitwiseUser{}
		}
	}

	users := []model.SplitwiseUser{}
	for _, user := range s._users {
		if user == nil {
			continue
		}
		users = append(users, *user)
	}
	return users
}

func (s *SplitwiseAPI) UpdateCurrencies() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s._lastCurrenciesDownload = time.Now()
	currencies, err := s._client.GetCurrencies(context.Background())
	if err != nil {
		return fmt.Errorf("error getting splitwise currencies: %s", err)
	}

	var currencyList []string
	for _, currency := range currencies {
		if currency.CurrencyCode == "" {
			continue
		}
		currencyList = append(currencyList, currency.CurrencyCode)
	}

	err = queries.UpdateSplitwiseConfigCurrencies(s.app.Db, currencyList)
	if err != nil {
		logging.Error("Error updating Splitwise currencies in database", "error", err)
		return fmt.Errorf("error updating Splitwise currencies in database: %w", err)
	}

	s._currencies = currencyList
	return nil
}

func (s *SplitwiseAPI) GetCurrencies() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	currencies := []string{}

	if len(s._currencies) == 0 && (time.Since(s._lastCurrenciesDownload) > 1*time.Minute || s._lastCurrenciesDownload.IsZero()) {
		logging.Debug("Updating Splitwise currencies")
		s.mutex.Unlock()
		err := s.UpdateCurrencies()
		s.mutex.Lock()
		if err != nil {
			logging.Error("Error updating Splitwise currencies", "error", err)
			return []string{}
		}
	}

	for _, currency := range s._currencies {
		if currency == "" {
			continue
		}
		currencies = append(currencies, currency)
	}
	return currencies
}

func (s *SplitwiseAPI) GetCurrency() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s._currency == "" {
		return "USD" // Default currency if not set
	}
	return s._currency
}

func (s *SplitwiseAPI) SetCurrency(currency string) error {
	s.mutex.Lock()
	if currency == "" || !slices.Contains(s._currencies, currency) {
		s.mutex.Unlock()
		return fmt.Errorf("invalid currency: %s", currency)
	}
	s.mutex.Unlock()

	err := queries.UpdateSplitwiseConfigInvoiceCurrency(s.app.Db, currency)
	if err != nil {
		logging.Error("Error updating Splitwise currency in database", "error", err)
		return fmt.Errorf("error updating Splitwise currency in database: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s._currency = currency
	return nil
}

func (s *SplitwiseAPI) Routes(route *gin.Engine, app *model.App) {
	splitwise := route.Group("/splitwise")
	{
		splitwise.GET("/", getSplitwiseConfigHandler(app, s))
		splitwise.POST("/", middleware.BlockIfNotAdmin(), postSplitwiseConfigHandler(app, s))
		splitwise.GET("/users", middleware.BlockIfNotAdmin(), getSplitwiseUsersHandler(app, s))
	}
}

func (s *SplitwiseAPI) WebLoader() templ.Component {
	return splitwise_view.Loader([]model.EB_EventType{model.EB_EVENT_INVOICE_CONFIG_UPDATE, model.EB_EVENT_NEW_BILL})
}
