package splitwise

import (
	"context"
	"fmt"
	"mc_manager/model"
	"os"
	"strconv"
	"sync"

	spw "github.com/aanzolaavila/splitwise.go"
)

// Check if the client has provided CLIENT_ID and CLIENT_SECRET
// check if there are old tokens to use
// check if the tokens are still valid
// if not, set not initialized
// if yes, set initialized

// if player

type SplitwiseAPI struct {
	app     *model.App
	token   string
	groupId int

	//protected
	mutex   sync.Mutex
	_client *spw.Client
	_state  model.InvoiceServiceStatus
}

func NewSplitwiseAPI(app *model.App) (*SplitwiseAPI, error) {
	splitwiseToken := os.Getenv("SPLITWISE_API_TOKEN")
	if splitwiseToken == "" {
		return nil, fmt.Errorf("SPLITWISE_GROUP_ID not set")
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
		app:     app,
		_client: nil,
		_state:  model.StatusInvoiceServiceNotInitialized,
		token:   splitwiseToken,
		groupId: splitwiseGroupId,
	}, nil
}

func (s *SplitwiseAPI) updateState(state model.InvoiceServiceStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s._state = state
	//TODO: update at central long poll queue
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

	s.updateState(model.StatusInvoiceServiceReady)
	return nil
}

func (s *SplitwiseAPI) Status() model.InvoiceServiceStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s._state
}

func (s *SplitwiseAPI) SyncBills() error {
	return nil
}
func (s *SplitwiseAPI) GetInvoice(bill *model.Bill) (*model.Invoice, error) {
	return nil, nil
}
func (s *SplitwiseAPI) CreateInvoice(bill *model.Bill) (*model.Invoice, error) {
	return nil, nil
}
func (s *SplitwiseAPI) GetUsers() ([]*model.InvoiceServiceUser, error) {
	s.mutex.Lock()
	if s._client == nil {
		s.mutex.Unlock()
		return nil, fmt.Errorf("splitwise client not initialized")
	}
	s.mutex.Unlock()

	group, err := s._client.GetGroup(context.Background(), s.groupId)
	if err != nil {
		return nil, fmt.Errorf("error getting splitwise group: %s", err)
	}

	var users []*model.InvoiceServiceUser
	for _, member := range group.Members {
		combinedNameValue := fmt.Sprintf("%s %s", member.FirstName, member.LastName)
		combinedName := &combinedNameValue
		if *combinedName == "" {
			combinedName = nil
		}

		users = append(users, &model.InvoiceServiceUser{
			Id:    fmt.Sprint(member.ID),
			Name:  combinedName,
			Email: &member.Email,
		})
	}
	return users, nil
}
