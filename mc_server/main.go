package main

import (
	"context"
	"log/slog"
	"os"
	"path"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

var Version = "development"
var logger *slog.Logger

func main() {
	if Version == "development" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logFile := &lumberjack.Logger{
			Filename:   "/ec2-user/logs/mc_server.log", // Set the filename where logs will be written
			MaxSize:    10,                             // Maximum size in megabytes before log rotation
			MaxBackups: 2,                              // Maximum number of old log files to retain
			MaxAge:     30,                             // Maximum number of days to retain old log file
		}
		defer logFile.Close()
		logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))

	}

	envPath := ".env"
	currentPath, err := os.Executable()
	if Version != "development" {
		if err != nil {
			logger.Error("Cant get Executable path", "error", err)
			os.Exit(1)
		}
		envPath = path.Join(path.Dir(currentPath), ".env")
	}
	if err := godotenv.Load(envPath); err != nil {
		logger.Error("Error loading .env file:", "error", err)
		os.Exit(1)
	}

	CONTAINER_NAME := os.Getenv("CONTAINER_NAME")
	if CONTAINER_NAME == "" {
		logger.Error("CONTAINER_NAME is not set in .env file")
		os.Exit(1)
	}

	StartContainer(CONTAINER_NAME)
	logger.Info("Container started", "name", CONTAINER_NAME)

	for {
		logger.Info("Checking if container is still running", "name", CONTAINER_NAME)
		isUp, err := IsContainerUp(CONTAINER_NAME)
		if err != nil {
			logger.Error("Error checking container status", "error", err)
			return
		}
		if !isUp {
			logger.Info("Container is not running, stopping updater")
			break
		}
	}

	//shutdown computer
	logger.Info("Container is not running, shutting down system")
}

func IsContainerUp(name string) (bool, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		logger.Info("Unable to create docker client")
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return false, err
	}

	logger.Info("Checking if container is already running")
	logger.Debug("Containers: ", "containers", containers)

	for _, container := range containers {
		if container.Names[0] == "/"+name {
			logger.Info("Container is already running")
			return true, nil
		}
	}
	return false, nil
}

func StartContainer(name string) error {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		logger.Info("Unable to create docker client")
		panic(err)
	}

	logger.Info("Starting container", "name", name)
	err = cli.ContainerStart(context.Background(), name, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}
