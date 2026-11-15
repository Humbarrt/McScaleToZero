package model

import "github.com/shopspring/decimal"

type Config interface {
	GetAWSRegion() string
	GetAWSAccessKeyId() string
	GetAWSSecretAccessKey() string
	GetAWSProjectName() string
	GetMCServerSSHUser() string
	GetMCServerSSHKey() []byte
	GetDashboardPort() int
	GetTransferServerPort() int
	GetInvoiceService() InvoiceServiceType
	GetVersion() string
	GetDataPath() string

	GetDashboardURL() string
	SetDashboardURL(url string)
	GetMCServerPort() int
	SetMCServerPort(port int) error
	GetMCOnlineImage() *string
	SetMCOnlineImage(image *string) error
	GetMCOnlineMOTD() []byte
	SetMCOnlineMOTD(motd []byte) error
	GetMCOfflineImage() *string
	SetMCOfflineImage(image *string) error
	GetMCOfflineMOTD() []byte
	SetMCOfflineMOTD(motd []byte) error
	GetReplaceMOTD() bool
	SetReplaceMOTD(replace bool) error
	GetLastImage() *string
	SetLastImage(image *string) error
	GetLastMOTD() []byte
	SetLastMOTD(motd []byte) error
	GetProtocolVersion() []byte
	SetProtocolVersion(version []byte) error
	GetMaxPlayers() int
	SetMaxPlayers(maxPlayers int) error
	GetBackupsToKeep() int
	SetBackupsToKeep(backupsToKeep int) error
	GetTaxRate() decimal.Decimal
	SetTaxRate(taxRate decimal.Decimal) error
	GetNetworkRatio() decimal.Decimal
	SetNetworkRatio(networkRatio decimal.Decimal) error
	GetFallbackFixedCost() decimal.Decimal
	SetFallbackFixedCost(fallbackFixedCost decimal.Decimal) error
	GetFallbackVariableCost() decimal.Decimal
	SetFallbackVariableCost(fallbackVariableCost decimal.Decimal) error
	GetLastInvoiceService() *InvoiceServiceType
}

type MutableConfig struct {
	MCServerPort         int
	OnlineImage          *string
	OnlineMOTD           []byte
	OfflineImage         *string
	OfflineMOTD          []byte
	ReplaceMOTD          bool
	LastImage            *string
	LastMOTD             []byte
	ProtocolVersion      *[]byte
	MaxPlayers           int
	BackupsToKeep        int
	TaxRate              decimal.Decimal
	NetworkRatio         decimal.Decimal
	FallbackFixedCost    decimal.Decimal
	FallbackVariableCost decimal.Decimal
	LastInvoiceService   *InvoiceServiceType
}
