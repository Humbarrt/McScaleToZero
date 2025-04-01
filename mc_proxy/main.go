package main

import (
	"log/slog"
	mc "mc_proxy/mc_net"
	"net"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

var isProduction = "false"

type Config struct {
	FLY_API_HOSTNAME string
	FLY_API_TOKEN    string
	FLY_APP_NAME     string
	FLY_MACHINE_ID   string
	MC_HOST          string
	PORT             string
	CHECK_INTERVAL   time.Duration
}

var IsOnline bool

var logger *slog.Logger

func init() {
	// Set up Lumberjack for log rotation
	logFile := &lumberjack.Logger{
		Filename:   "/root/logs/mc-proxy.log",
		MaxSize:    10,   // MB
		MaxBackups: 3,    // Keep last 3 log files
		MaxAge:     28,   // Days
		Compress:   true, // Compress old logs
	}

	// Create a new slog logger with a text handler
	if isProduction == "true" {
		logger = slog.New(
			slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	} else {
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}
}

// Handle incoming player connections
func handleConnection(clientConn net.Conn, config *Config) {
	logger.Info("Player attempting to connect...", "remoteAddr", clientConn.RemoteAddr().String())
	handshake, err := mc.ReadHandshake(clientConn)
	if err != nil {
		logger.Error("Error reading handshake:", "error", err)
		clientConn.Close()
		return
	}
	logger.Info("Player handshake received", "handshake", handshake)
	if handshake.NextState == 2 {
		logger.Info("Player is connecting to the server", "remoteAddr", clientConn.RemoteAddr().String())

		//Wait for Login packet
		loginStart, err := mc.ReadLoginStart(clientConn)
		if err != nil {
			logger.Error("Error reading login start:", "error", err)
			clientConn.Close()
			return
		}
		logger.Info("Player is logging in", "username", loginStart.Username, "uuid", loginStart.UUID)
		if loginStart.UUID == nil {
			//TODO: request the uuid from mojang
			logger.Error("UUID is nil, requesting from mojang")
		}

		uuid := *loginStart.UUID
		// Send LoginSuccess packet
		loginSuccess := mc.WriteLoginSuccess(clientConn, uuid, loginStart.Username)
		if loginSuccess != nil {
			logger.Error("Error writing login success:", "error", loginSuccess)
			clientConn.Close()
			return
		}
		logger.Info("Sent LoginSuccess", "username", loginStart.Username)

		// Wait for LoginAcknowledged packet
		loginAcknowledged, err := mc.ReadLoginAcknowledged(clientConn)
		if err != nil {
			logger.Error("Error reading login acknowledged:", "error", loginAcknowledged)
			clientConn.Close()
			return
		}
		logger.Info("Player acknowledged login", "loggedIn", loginAcknowledged, "username", loginStart.Username)

		// Send Transfer packet
		err = mc.WriteTransfer(clientConn, "org.playskyward.gg", 25565)
		if err != nil {
			logger.Error("Error writing transfer:", "error", err)
			clientConn.Close()
			return
		}
		logger.Info("Sent Transfer packet", "host", "org.playskyward.gg", "port", 25565)
	}
}
func main() {
	currentPath, err := os.Executable()
	if err != nil {
		logger.Error("Cant get Executable path", "error", err)
		os.Exit(1)
	}
	logger.Debug("Is production:", "isProduction", isProduction)

	envPath := ".env"
	if isProduction == "true" {
		envPath = path.Join(path.Dir(currentPath), ".env")
	}
	if err := godotenv.Load(envPath); err != nil {
		logger.Error("Error loading .env file:", "error", err)
		os.Exit(1)
	}

	checkInterval, err := time.ParseDuration(os.Getenv("CHECK_INTERVAL"))
	if err != nil {
		logger.Error("Error parsing CHECK_INTERVAL:", "error", err)
		os.Exit(1)
	}
	config := &Config{
		FLY_API_HOSTNAME: os.Getenv("FLY_API_HOSTNAME"),
		FLY_API_TOKEN:    os.Getenv("FLY_API_TOKEN"),
		FLY_APP_NAME:     os.Getenv("FLY_APP_NAME"),
		FLY_MACHINE_ID:   os.Getenv("FLY_MACHINE_ID"),
		MC_HOST:          os.Getenv("MC_HOST"),
		PORT:             os.Getenv("PORT"),
		CHECK_INTERVAL:   checkInterval,
	}

	v := reflect.ValueOf(*config)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "" {
			logger.Error("Error: missing environment variable", "env", v.Type().Field(i).Name)
			os.Exit(1)
		}
	}

	listener, err := net.Listen("tcp", ":"+config.PORT)
	if err != nil {
		logger.Error("Failed to start proxy:", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("TCP Proxy listening on port " + config.PORT + "...")

	for {
		logger.Debug("Waiting for incoming connections...")
		clientConn, err := listener.Accept()
		if err != nil {
			logger.Error("Connection failed:", "error", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(clientConn, config)
	}
}
