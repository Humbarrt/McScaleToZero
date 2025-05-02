package none

import (
	"mc_analyser/model"
)

type NoInvoiceService struct {
}

func NewNoInvoiceService(app *model.App) *NoInvoiceService {
	return &NoInvoiceService{}
}

func (s *NoInvoiceService) Init() error {
	return nil
}

func (s *NoInvoiceService) Status() model.InvoiceServiceStatus {
	return model.StatusInvoiceServiceNotInitialized
}

func (s *NoInvoiceService) SyncBills() error {
	return nil
}
func (s *NoInvoiceService) GetUsers() ([]*model.InvoiceServiceUser, error) {
	return []*model.InvoiceServiceUser{}, nil
}
