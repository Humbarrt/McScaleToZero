package config

import (
	"mc_analyser/logging"
	"mc_analyser/model"
	"os"
	"path"
	"strconv"

	"github.com/joho/godotenv"
)

var isProduction = "false"

func Load() *model.Config {
	logging.Info("Version", "isProduction", isProduction)
	currentPath, err := os.Executable()
	if err != nil {
		logging.Fatal("Cant get Executable path", "error", err)
	}

	envPath := ".env"
	if isProduction == "true" {
		envPath = path.Join(path.Dir(currentPath), ".env")
	}
	if err := godotenv.Load(envPath); err != nil {
		logging.Fatal("Error loading .env file:", "error", err)
	}

	mcPortString := os.Getenv("MC_PORT")
	if mcPortString == "" {
		logging.Fatal("MC_PORT is not set in the environment variables")
	}
	mcPort, err := strconv.Atoi(mcPortString)
	if err != nil {
		logging.Fatal("Error converting MC_PORT to int:", "error", err)
	}

	serverWatcherRetriesString := os.Getenv("SERVER_WATCHER_RETRIES")
	if serverWatcherRetriesString == "" {
		logging.Fatal("SERVER_WATCHER_RETRIES is not set in the environment variables")
	}
	serverWatcherRetries, err := strconv.Atoi(serverWatcherRetriesString)
	if err != nil {
		logging.Fatal("Error converting SERVER_WATCHER_RETRIES to int:", "error", err)
	}

	config := &model.Config{
		AWS_ACCESS_KEY_ID:      os.Getenv("AWS_ACCESS_KEY_ID"),
		AWS_SECRET_ACCESS_KEY:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWS_INSTANCE_NAME:      os.Getenv("AWS_INSTANCE_NAME"),
		MC_PORT:                mcPort,
		SERVER_WATCHER_RETRIES: serverWatcherRetries,
	}

	for _, env := range []string{config.AWS_ACCESS_KEY_ID, config.AWS_SECRET_ACCESS_KEY, config.AWS_INSTANCE_NAME} {
		if env == "" {
			logging.Fatal("Environment variable is not set", "variable", env)
		}
	}

	//TODO: splitwise token
	//TODO: splitwise group id
	//TODO: PROMETHEUS_URL
	//TODO: PROMETHEUS_ORG_ID
	//TODO: MC SERVER IP
	logging.Debug("Config loaded")
	return config
}
