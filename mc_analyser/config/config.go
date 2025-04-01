package config

import (
	"mc_analyser/logging"
	"mc_analyser/model"
	"os"
	"path"
	"reflect"

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

	config := &model.Config{
		FLY_API_BASE_URL: os.Getenv("FLY_API_BASE_URL"),
		FLY_API_TOKEN:    os.Getenv("FLY_API_TOKEN"),
		FLY_APP_NAME:     os.Getenv("FLY_APP_NAME"),
		SIGNING_KEY:      []byte(os.Getenv("SIGNING_KEY")),
	}

	v := reflect.ValueOf(*config)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "" {
			logging.Fatal("Error: missing environment variable", "env", v.Type().Field(i).Name)
		}
	}

	//TODO: splitwise token
	//TODO: splitwise group id
	//TODO: PROMETHEUS_URL
	//TODO: PROMETHEUS_ORG_ID
	//TODO: MC SERVER IP

	return config
}
