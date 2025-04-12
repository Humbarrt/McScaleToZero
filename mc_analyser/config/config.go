package config

import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"mc_analyser/db"
	"mc_analyser/db/queries"
	"mc_analyser/logging"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var version = "development"

type config struct {
	awsAccessKeyId     string
	awsSecretAccessKey string
	awsInstanceName    string
	splitwiseApiKey    string
	splitwiseGroupId   string
	mcServerSSHUser    string
	mcServerSSHKeyFile string
	mcServerSSHKey     []byte
	dataPath           string
	version            string
	db                 *db.DB

	//protected
	mutex           sync.Mutex
	_mcPort         int
	_mcOfflineImage *[]byte
	_mcOfflineMOTD  *string
	_ipCheck        bool
	_backupsToKeep  int
}

func (c *config) GetAWSAccessKeyId() string {
	return c.awsAccessKeyId
}
func (c *config) GetAWSSecretAccessKey() string {
	return c.awsSecretAccessKey
}
func (c *config) GetAWSInstanceName() string {
	return c.awsInstanceName
}
func (c *config) GetSplitwiseApiKey() string {
	return c.splitwiseApiKey
}
func (c *config) GetSplitwiseGroupId() string {
	return c.splitwiseGroupId
}
func (c *config) GetMCServerSSHUser() string {
	return c.mcServerSSHUser
}
func (c *config) GetMCServerSSHKey() []byte {
	return c.mcServerSSHKey
}
func (ch *config) GetVersion() string {
	return ch.version
}
func (c *config) GetDataPath() string {
	return c.dataPath
}

func (ch *config) IsDevMode() bool {
	return ch.GetVersion() == "development"
}

func (c *config) GetMCPort() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c._mcPort
}
func (c *config) SetMCPort(port int) error {
	err := queries.UpdateMCServerPort(c.db, port)
	if err != nil {
		logging.Error("Error updating MC server port in database:", err)
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c._mcPort = port
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

	dynConfig, err := queries.GetConfig(db)
	if err != nil || dynConfig == nil {
		logging.Fatal("Error getting config", "error", err, "config", dynConfig)
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = path.Join(path.Dir(currentPath), "data")
	}
	mcServerSSHUser := os.Getenv("MC_SERVER_SSH_USER")
	if mcServerSSHUser == "" {
		mcServerSSHUser = "ec2-user"
	}
	mcServerSSHKeyFile := os.Getenv("MC_SERVER_SSH_KEY_FILE")
	if mcServerSSHKeyFile == "" {
		mcServerSSHKeyFile = "mc-server-key.pem"
	}
	mcServerSSHKey, err := os.ReadFile(path.Join(dataPath, mcServerSSHKeyFile))
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	if len(mcServerSSHKey) == 0 {
		log.Fatalf("private key is empty")
	}

	config := &config{
		awsAccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
		awsSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		awsInstanceName:    os.Getenv("AWS_INSTANCE_NAME"),
		splitwiseApiKey:    os.Getenv("SPLITWISE_API_KEY"),
		splitwiseGroupId:   os.Getenv("SPLITWISE_GROUP_ID"),
		mcServerSSHUser:    mcServerSSHUser,
		mcServerSSHKeyFile: mcServerSSHKeyFile,
		mcServerSSHKey:     mcServerSSHKey,
		dataPath:           dataPath,
		db:                 db,
		_mcPort:            dynConfig.MCServerPort,
		_mcOfflineImage:    dynConfig.OfflineImage,
		_mcOfflineMOTD:     dynConfig.OfflineMOTD,
		_ipCheck:           dynConfig.IPCheck,
		_backupsToKeep:     dynConfig.BackupsToKeep,
		version:            version,
	}

	for _, env := range []string{config.awsAccessKeyId, config.awsSecretAccessKey, config.awsInstanceName, config.splitwiseApiKey, config.splitwiseGroupId, config.mcServerSSHUser, config.mcServerSSHKeyFile} {
		if env == "" {
			logging.Fatal("Environment variable is not set", "variable", env)
		}
	}

	logging.Debug("Config loaded")
	return config
}
