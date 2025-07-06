package config

import (
	"fmt"
	"mc_manager/db"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

var version = "development"

type config struct {
	awsRegion          string
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsProjectName     string
	mcServerSSHUser    string
	mcServerSSHKeyFile string
	mcServerSSHKey     []byte
	dashboardPort      int
	transferServerPort int
	invoiceService     model.InvoiceServiceType
	dataPath           string
	version            string
	db                 *db.DB

	//protected
	mutex                 sync.Mutex
	_dashboardURL         string
	_mcServerPort         int
	_mcOnlineImage        *string
	_mcOnlineMOTD         []byte
	_mcOfflineImage       *string
	_mcOfflineMOTD        []byte
	_replaceMOTD          bool
	_lastImage            *string
	_lastMOTD             []byte
	_protocolVersion      *[]byte
	_maxPlayers           int
	_backupsToKeep        int
	_taxRate              decimal.Decimal
	_networkRatio         decimal.Decimal
	_fallbackFixedCost    decimal.Decimal
	_fallbackVariableCost decimal.Decimal
}

func (c *config) GetAWSRegion() string {
	return c.awsRegion
}

func (c *config) GetAWSAccessKeyId() string {
	return c.awsAccessKeyId
}
func (c *config) GetAWSSecretAccessKey() string {
	return c.awsSecretAccessKey
}
func (c *config) GetAWSProjectName() string {
	return c.awsProjectName
}
func (c *config) GetMCServerSSHUser() string {
	return c.mcServerSSHUser
}
func (c *config) GetMCServerSSHKey() []byte {
	return c.mcServerSSHKey
}
func (c *config) GetDashboardPort() int {
	return c.dashboardPort
}
func (c *config) GetTransferServerPort() int {
	return c.transferServerPort
}
func (c *config) GetInvoiceService() model.InvoiceServiceType {
	return c.invoiceService
}
func (c *config) GetDataPath() string {
	return c.dataPath
}
func (ch *config) GetVersion() string {
	return ch.version
}
func (ch *config) IsDevMode() bool {
	return ch.GetVersion() == "development"
}

func (c *config) GetDashboardURL() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c._dashboardURL == "" {
		c._dashboardURL = fmt.Sprintf("http://localhost:%d", c.dashboardPort)
	}
	return c._dashboardURL
}
func (c *config) SetDashboardURL(url string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._dashboardURL = url
}

func (c *config) GetMCServerPort() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcServerPort
}
func (c *config) SetMCServerPort(port int) error {
	err := queries.UpdateMCServerPort(c.db, port)
	if err != nil {
		logging.Error("Error updating MC server port in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcServerPort = port
	return nil
}

func (c *config) GetMCOnlineImage() *string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOnlineImage
}
func (c *config) SetMCOnlineImage(base64Image *string) error {
	if base64Image != nil && len(*base64Image) != 0 {
		err := ValidateImage(base64Image)
		if err != nil {
			logging.Error("Invalid image:", "error", err)
			return err
		}
	}

	err := queries.UpdateOnlineImage(c.db, base64Image)
	if err != nil {
		logging.Error("Error updating offline image in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOnlineImage = base64Image
	return nil
}

func (c *config) GetMCOnlineMOTD() []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOnlineMOTD
}
func (c *config) SetMCOnlineMOTD(motd []byte) error {
	err := queries.UpdateOnlineMOTD(c.db, motd)
	if err != nil {
		logging.Error("Error updating offline MOTD in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOnlineMOTD = motd
	return nil
}

func (c *config) GetMCOfflineImage() *string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOfflineImage
}
func (c *config) SetMCOfflineImage(base64Image *string) error {
	if base64Image != nil && len(*base64Image) != 0 {
		err := ValidateImage(base64Image)
		if err != nil {
			logging.Error("Invalid image:", "error", err)
			return err
		}
	}

	err := queries.UpdateOfflineImage(c.db, base64Image)
	if err != nil {
		logging.Error("Error updating offline image in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOfflineImage = base64Image
	return nil
}

func (c *config) GetMCOfflineMOTD() []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOfflineMOTD
}
func (c *config) SetMCOfflineMOTD(motd []byte) error {
	err := queries.UpdateOfflineMOTD(c.db, motd)
	if err != nil {
		logging.Error("Error updating offline MOTD in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOfflineMOTD = motd
	return nil
}

func (c *config) GetReplaceMOTD() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._replaceMOTD
}
func (c *config) SetReplaceMOTD(replaceMOTD bool) error {
	err := queries.UpdateReplaceMOTD(c.db, replaceMOTD)
	if err != nil {
		logging.Error("Error updating replace MOTD in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._replaceMOTD = replaceMOTD
	return nil
}

func (c *config) GetLastImage() *string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._lastImage
}
func (c *config) SetLastImage(base64Image *string) error {
	if base64Image != nil && len(*base64Image) != 0 {
		err := ValidateImage(base64Image)
		if err != nil {
			logging.Error("Invalid image:", "error", err)
			return err
		}
	}

	err := queries.UpdateLastImage(c.db, base64Image)
	if err != nil {
		logging.Error("Error updating last image in database:", "error", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._lastImage = base64Image
	return nil
}

func (c *config) GetLastMOTD() []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._lastMOTD
}
func (c *config) SetLastMOTD(motd []byte) error {
	err := queries.UpdateLastMOTD(c.db, motd)
	if err != nil {
		logging.Error("Error updating last MOTD in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._lastMOTD = motd
	return nil
}

func (c *config) GetProtocolVersion() []byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return *c._protocolVersion
}
func (c *config) SetProtocolVersion(protocolVersion []byte) error {
	err := queries.UpdateProtocolVersion(c.db, &protocolVersion)
	if err != nil {
		logging.Error("Error updating protocol version in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._protocolVersion = &protocolVersion
	return nil
}

func (c *config) GetMaxPlayers() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._maxPlayers
}
func (c *config) SetMaxPlayers(maxPlayers int) error {
	err := queries.UpdateMaxPlayers(c.db, maxPlayers)
	if err != nil {
		logging.Error("Error updating max players in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._maxPlayers = maxPlayers
	return nil
}

func (c *config) GetBackupsToKeep() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._backupsToKeep
}
func (c *config) SetBackupsToKeep(backupsToKeep int) error {
	err := queries.UpdateBackupsToKeep(c.db, backupsToKeep)
	if err != nil {
		logging.Error("Error updating backups to keep in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._backupsToKeep = backupsToKeep
	return nil
}

func (c *config) GetTaxRate() decimal.Decimal {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._taxRate
}
func (c *config) SetTaxRate(taxRate decimal.Decimal) error {
	err := queries.UpdateTaxRate(c.db, taxRate)
	if err != nil {
		logging.Error("Error updating tax rate in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._taxRate = taxRate
	return nil
}

func (c *config) GetNetworkRatio() decimal.Decimal {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._networkRatio
}
func (c *config) SetNetworkRatio(networkRatio decimal.Decimal) error {
	err := queries.UpdateNetworkRatio(c.db, networkRatio)
	if err != nil {
		logging.Error("Error updating tax rate in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._networkRatio = networkRatio
	return nil
}

func (c *config) GetFallbackFixedCost() decimal.Decimal {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._fallbackFixedCost
}
func (c *config) SetFallbackFixedCost(fallbackFixedCost decimal.Decimal) error {
	err := queries.UpdateFallbackFixedCost(c.db, fallbackFixedCost)
	if err != nil {
		logging.Error("Error updating tax rate in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._fallbackFixedCost = fallbackFixedCost
	return nil
}

func (c *config) GetFallbackVariableCost() decimal.Decimal {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._fallbackVariableCost
}
func (c *config) SetFallbackVariableCost(fallbackVariableCost decimal.Decimal) error {
	err := queries.UpdateFallbackVariableCost(c.db, fallbackVariableCost)
	if err != nil {
		logging.Error("Error updating tax rate in database:", "error", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._fallbackVariableCost = fallbackVariableCost
	return nil
}

func NewConfig() *config {
	logging.Info("Version", "version", version)
	currentPath, err := os.Executable()
	if err != nil {
		logging.Fatal("Cant get Executable path", "error", err)
	}

	envPath := ".env"
	if version != "development" {
		envPath = path.Join(path.Dir(currentPath), ".env")
	}
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		if err := godotenv.Load(envPath); err != nil {
			logging.Fatal("Error loading .env file:", "error", err)
		}
	}

	dataPath := getenv("DATA_PATH", path.Join(path.Dir(currentPath), "data"))
	return &config{
		dataPath: dataPath,
		version:  version,
	}
}

func (c *config) Load(db *db.DB) {
	mutConfig, err := queries.GetMutableConfig(db)
	if err != nil || mutConfig == nil {
		logging.Fatal("Error getting config", "error", err, "config", mutConfig)
	}

	mcServerSSHUser := getenv("MC_SERVER_SSH_USER", "ec2-user")
	mcServerSSHKeyFile := getenv("MC_SERVER_SSH_KEY_FILE", "aws-server-key.pem")
	mcServerSSHKey, err := os.ReadFile(path.Join(c.dataPath, mcServerSSHKeyFile))
	if err != nil {
		logging.Fatal("unable to read private key:", "error", err)
	} else if len(mcServerSSHKey) == 0 {
		logging.Fatal("private key is empty")
	}

	awsRegion := getenv("AWS_REGION", "")
	awsAccessKeyId := getenv("AWS_ACCESS_KEY_ID", "")
	awsSecretAccessKey := getenv("AWS_SECRET_ACCESS_KEY", "")
	awsRegionFile := getenv("AWS_REGION_FILE", "aws-region.txt")
	awsAccessKeyIdFile := getenv("AWS_ACCESS_KEY_ID_FILE", "aws-access-key-id.txt")
	awsSecretAccessKeyFile := getenv("AWS_SECRET_ACCESS_KEY_FILE", "aws-secret-access-key.txt")

	if awsRegion != "" && awsAccessKeyId != "" && awsSecretAccessKey != "" {
		logging.Debug("Using AWS credentials from environment variables")
	} else if awsRegion == "" && awsAccessKeyId == "" && awsSecretAccessKey == "" {
		logging.Debug("Using AWS credentials from file")

		awsRegionBytes, err := os.ReadFile(path.Join(c.dataPath, awsRegionFile))
		if err != nil {
			logging.Fatal("unable to read AWS region file:", "error", err)
		} else if len(awsRegionBytes) == 0 {
			logging.Fatal("AWS region file is empty")
		}
		awsRegion = string(awsRegionBytes)

		awsAccessKeyIdBytes, err := os.ReadFile(path.Join(c.dataPath, awsAccessKeyIdFile))
		if err != nil {
			logging.Fatal("unable to read AWS access key ID file:", "error", err)
		} else if len(awsAccessKeyIdBytes) == 0 {
			logging.Fatal("AWS access key ID file is empty")
		}
		awsAccessKeyId = string(awsAccessKeyIdBytes)

		awsSecretAccessKeyBytes, err := os.ReadFile(path.Join(c.dataPath, awsSecretAccessKeyFile))
		if err != nil {
			logging.Fatal("unable to read AWS secret access key file:", "error", err)
		} else if len(awsSecretAccessKeyBytes) == 0 {
			logging.Fatal("AWS secret access key file is empty")
		}

		awsSecretAccessKey = string(awsSecretAccessKeyBytes)
	} else {
		logging.Fatal("You need to provide AWS credentials. Either put them in DATA_PATH and specify AWS_REGION_FILE & AWS_ACCESS_KEY_ID_FILE & AWS_SECRET_ACCESS_KEY_FILE or set them in environment variables with AWS_REGION & AWS_ACCESS_KEY_ID & AWS_SECRET_ACCESS_KEY")
	}

	awsProjectName := getenv("AWS_PROJECT_NAME", "minecraft")
	dashboardPortString := getenv("DASHBOARD_PORT", "8080")
	dashboardPort, err := strconv.Atoi(dashboardPortString)
	if err != nil {
		logging.Fatal("Error parsing dashboard port:", "error", err)
	} else if dashboardPort <= 0 || dashboardPort > 65535 {
		logging.Fatal("Invalid dashboard port:", "port", dashboardPort)
	}

	transferServerPortString := getenv("TRANSFER_SERVER_PORT", "25565")
	transferServerPort, err := strconv.Atoi(transferServerPortString)
	if err != nil {
		logging.Fatal("Error parsing transfer server port:", "error", err)
	} else if transferServerPort <= 0 || transferServerPort > 65535 {
		logging.Fatal("Invalid transfer server port:", "port", transferServerPort)
	}

	//Check if invoice service is set
	invoiceServiceString := strings.ToLower(getenv("INVOICE_SERVICE", string(model.InvoiceServiceNone)))
	var invoiceService model.InvoiceServiceType
	switch invoiceServiceString {
	case string(model.InvoiceServiceNone):
		invoiceService = model.InvoiceServiceNone
		logging.Debug("Invoice service is set to none")
	case string(model.InvoiceServiceSplitwise):
		invoiceService = model.InvoiceServiceSplitwise
		logging.Debug("Invoice service is set to splitwise")
	default:
		logging.Fatal("Invalid invoice service:", "service", invoiceService)
	}

	c.awsRegion = awsRegion
	c.awsAccessKeyId = awsAccessKeyId
	c.awsSecretAccessKey = awsSecretAccessKey
	c.awsProjectName = awsProjectName
	c.mcServerSSHUser = mcServerSSHUser
	c.mcServerSSHKeyFile = mcServerSSHKeyFile
	c.mcServerSSHKey = mcServerSSHKey
	c.dashboardPort = dashboardPort
	c.transferServerPort = transferServerPort
	c.invoiceService = invoiceService
	c.version = version
	c.db = db

	c._mcServerPort = mutConfig.MCServerPort
	c._mcOnlineImage = mutConfig.OnlineImage
	c._mcOnlineMOTD = mutConfig.OnlineMOTD
	c._mcOfflineImage = mutConfig.OfflineImage
	c._mcOfflineMOTD = mutConfig.OfflineMOTD
	c._replaceMOTD = mutConfig.ReplaceMOTD
	c._lastImage = mutConfig.LastImage
	c._lastMOTD = mutConfig.LastMOTD
	c._protocolVersion = mutConfig.ProtocolVersion
	c._maxPlayers = mutConfig.MaxPlayers
	c._backupsToKeep = mutConfig.BackupsToKeep
	c._taxRate = mutConfig.TaxRate
	c._networkRatio = mutConfig.NetworkRatio
	c._fallbackFixedCost = mutConfig.FallbackFixedCost
	c._fallbackVariableCost = mutConfig.FallbackVariableCost

	logging.Debug("Config loaded")
}
