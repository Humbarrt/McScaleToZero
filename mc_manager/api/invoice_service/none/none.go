package none

import (
	"mc_manager/model"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
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

func (s *NoInvoiceService) Routes(route *gin.RouterGroup, app *model.App) {
}

func (s *NoInvoiceService) WebLoader() templ.Component {
	return templ.NopComponent
}
