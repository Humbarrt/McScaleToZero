package config

import (
	"encoding/base64"
	"fmt"
	"image"
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
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsProjectName     string
	mcServerSSHUser    string
	mcServerSSHKeyFile string
	mcServerSSHKey     []byte
	dashboardPort      int
	transferServerPort int
	invoiceService     model.InvoiceService
	dataPath           string
	version            string
	db                 *db.DB

	//protected
	mutex           sync.Mutex
	_dashboardURL   string
	_mcServerPort   int
	_mcOfflineImage *[]byte
	_mcOfflineMOTD  *string
	_ipCheck        bool
	_backupsToKeep  int
	_taxRate        decimal.Decimal
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
func (c *config) GetInvoiceService() model.InvoiceService {
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
		logging.Error("Error updating MC server port in database:", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcServerPort = port
	return nil
}

func (c *config) GetMCOfflineImage() *[]byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOfflineImage
}
func (c *config) SetMCOfflineImage(base64Image *[]byte) error {
	if base64Image != nil {
		//Validate Base64
		base64ImageStr := string(*base64Image)
		if strings.HasPrefix(base64ImageStr, "data:") {
			parts := strings.SplitN(base64ImageStr, ",", 2)
			if len(parts) == 2 {
				base64ImageStr = parts[1]
			}
		}

		// Decode Base64
		imgData, err := base64.StdEncoding.DecodeString(base64ImageStr)
		if err != nil {
			logging.Error("Error decoding Base64:", err)
			return err
		}

		// Decode image
		img, _, err := image.Decode(strings.NewReader(string(imgData)))
		if err != nil {
			logging.Error("Error decoding image:", err)
			return err
		}

		// Get image bounds
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		fmt.Printf("Width: %d, Height: %d\n", width, height)
		if width != 64 || height != 64 {
			logging.Error("Image dimensions are not 64x64")
			return fmt.Errorf("image dimensions are not 64x64")
		}
	}

	err := queries.UpdateOfflineImage(c.db, base64Image)
	if err != nil {
		logging.Error("Error updating offline image in database:", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOfflineImage = base64Image
	return nil
}

func (c *config) GetMCOfflineMOTD() *string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcOfflineMOTD
}
func (c *config) SetMCOfflineMOTD(motd *string) error {
	err := queries.UpdateOfflineMOTD(c.db, motd)
	if err != nil {
		logging.Error("Error updating offline MOTD in database:", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcOfflineMOTD = motd
	return nil
}

func (c *config) GetIPCheck() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._ipCheck
}
func (c *config) SetIPCheck(ipCheck bool) error {
	err := queries.UpdateIPCheck(c.db, ipCheck)
	if err != nil {
		logging.Error("Error updating MC server port in database:", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._ipCheck = ipCheck
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
		logging.Error("Error updating backups to keep in database:", err)
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
		logging.Error("Error updating tax rate in database:", err)
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._taxRate = taxRate
	return nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func Load(db *db.DB) *config {
	logging.Info("Version", "version", version)
	currentPath, err := os.Executable()
	if err != nil {
		logging.Fatal("Cant get Executable path", "error", err)
	}

	envPath := ".env"
	if version != "development" {
		envPath = path.Join(path.Dir(currentPath), ".env")
	}
	if err := godotenv.Load(envPath); err != nil {
		logging.Fatal("Error loading .env file:", "error", err)
	}

	mutConfig, err := queries.GetMutableConfig(db)
	if err != nil || mutConfig == nil {
		logging.Fatal("Error getting config", "error", err, "config", mutConfig)
	}

	dataPath := getenv("DATA_PATH", path.Join(path.Dir(currentPath), "data"))
	mcServerSSHUser := getenv("MC_SERVER_SSH_USER", "ec2-user")
	mcServerSSHKeyFile := getenv("MC_SERVER_SSH_KEY_FILE", "mc-server-key.pem")
	mcServerSSHKey, err := os.ReadFile(path.Join(dataPath, mcServerSSHKeyFile))
	if err != nil {
		logging.Fatal("unable to read private key:", "error", err)
	} else if len(mcServerSSHKey) == 0 {
		logging.Fatal("private key is empty")
	}

	awsAccessKeyId := getenv("AWS_ACCESS_KEY_ID", "")
	awsSecretAccessKey := getenv("AWS_SECRET_ACCESS_KEY", "")
	awsAccessKeyIdFile := getenv("AWS_ACCESS_KEY_ID_FILE", "mc-access-key-id.txt")
	awsSecretAccessKeyFile := getenv("AWS_SECRET_ACCESS_KEY_FILE", "mc-secret-access-key.txt")

	if awsAccessKeyId != "" && awsSecretAccessKey != "" {
		logging.Debug("Using AWS credentials from environment variables")
	} else if awsAccessKeyId == "" && awsSecretAccessKey == "" {
		logging.Debug("Using AWS credentials from file")

		awsAccessKeyIdBytes, err := os.ReadFile(path.Join(dataPath, awsAccessKeyIdFile))
		if err != nil {
			logging.Fatal("unable to read AWS access key ID file:", "error", err)
		} else if len(awsAccessKeyIdBytes) == 0 {
			logging.Fatal("AWS access key ID file is empty")
		}
		awsAccessKeyId = string(awsAccessKeyIdBytes)

		awsSecretAccessKeyBytes, err := os.ReadFile(path.Join(dataPath, awsSecretAccessKeyFile))
		if err != nil {
			logging.Fatal("unable to read AWS secret access key file:", "error", err)
		} else if len(awsSecretAccessKeyBytes) == 0 {
			logging.Fatal("AWS secret access key file is empty")
		}

		awsSecretAccessKey = string(awsSecretAccessKeyBytes)
	} else {
		logging.Fatal("You need to provide AWS credentials. Either put them in DATA_PATH and specify AWS_ACCESS_KEY_ID_FILE & AWS_SECRET_ACCESS_KEY_FILE or set them in environment variables with AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
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
	var invoiceService model.InvoiceService
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

	config := &config{
		awsAccessKeyId:     awsAccessKeyId,
		awsSecretAccessKey: awsSecretAccessKey,
		awsProjectName:     awsProjectName,
		mcServerSSHUser:    mcServerSSHUser,
		mcServerSSHKeyFile: mcServerSSHKeyFile,
		mcServerSSHKey:     mcServerSSHKey,
		dashboardPort:      dashboardPort,
		transferServerPort: transferServerPort,
		invoiceService:     invoiceService,
		dataPath:           dataPath,
		version:            version,
		db:                 db,

		_mcServerPort:   mutConfig.MCServerPort,
		_mcOfflineImage: mutConfig.OfflineImage,
		_mcOfflineMOTD:  mutConfig.OfflineMOTD,
		_ipCheck:        mutConfig.IPCheck,
		_backupsToKeep:  mutConfig.BackupsToKeep,
		_taxRate:        mutConfig.TaxRate,
	}

	logging.Debug("Config loaded")
	return config
}
