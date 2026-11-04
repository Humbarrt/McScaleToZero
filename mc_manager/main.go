package main

import (
	"crypto/rand"
	"embed"
	"mc_manager/api/aws"
	"mc_manager/api/invoice_service/none"
	"mc_manager/api/invoice_service/splitwise"
	"mc_manager/db"
	"mc_manager/logging"
	"mc_manager/model"
	"mc_manager/web/routes"
	"time"

	_config "mc_manager/config"
	cw "mc_manager/processes/cost_watcher"
	"mc_manager/processes/decoy_server"
	eb "mc_manager/processes/event_broadcaster"
	lw "mc_manager/processes/log_watcher"
	sw "mc_manager/processes/server_watcher"
	sm "mc_manager/processes/session_manager"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

//go:embed web/static/*
var fs embed.FS

/* func test() error {
	backupFilePath := "/Users/humpert/Projects/McServer/world.zip"
	_, backupFileName := path.Split(backupFilePath)
	read, err := zip.OpenReader(backupFilePath)
	if err != nil {
		logging.Error("Error reading world zip", "error", err)
		return err
	}
	defer read.Close()

	backupFileName = strings.Split(backupFileName, ".")[0]
	statsPath := path.Join(backupFileName, "stats")
	logging.Info("Searching for", "path", statsPath)
	for _, f := range read.File {
		currPath, currName := path.Split(f.Name)
		if (currPath != statsPath && currPath != statsPath+"/") || currName == "" {
			continue
		}
		logging.Info("FOUND STATS FILE", "path", currPath, "name", currName)

		playerUUID := strings.Split(currName, ".")
		queries.Get

		v, err := f.Open()
		if err != nil {
			logging.Error("Error reading stat file from zip", "error", err)
			return err
		}
		defer v.Close()

	}

	return nil
} */

func main() {
	config := _config.NewConfig()
	logging.Info("Starting MC Manager")
	database, err := db.NewDatabase(config.GetDataPath())
	if err != nil {
		logging.Fatal("Error creating database connection", "error", err)
	}
	config.Load(database)

	//generate jwt secret
	secret := make([]byte, 64)
	_, err = rand.Read(secret)
	if err != nil {
		logging.Fatal("Error creating JWTSecret", "error", err)
	}

	router := gin.Default()
	app := model.App{
		Db:         database,
		Config:     config,
		Router:     router,
		FS:         fs,
		WebTimeout: 10 * time.Second,
		JWTSecret:  secret,
		SW:         nil,
		CW:         nil,
		IS:         nil,
		LW:         nil,
	}

	switch app.Config.GetInvoiceService() {
	case model.InvoiceServiceSplitwise:
		logging.Info("Invoice service configured to Splitwise")
		app.IS, err = splitwise.NewSplitwiseAPI(&app)
		if err != nil {
			logging.Fatal("Error creating invoice service", "error", err)
		}
		err = app.IS.Init()
		if err != nil {
			logging.Fatal("Error initializing invoice service", "error", err)
		}
	case model.InvoiceServiceNone:
		logging.Info("Invoice service not configured")
		app.IS = none.NewNoInvoiceService(&app)
	default:
		logging.Fatal("Unknown invoice service")
	}

	app.AWS, err = aws.NewAWSAPI(&app, "Project")
	if err != nil {
		logging.Fatal("Error creating AWS API", "error", err)
	}

	app.EB = eb.NewEventBroadcaster()

	app.SW = sw.NewServerWatcher(&app)
	go app.SW.StartWatchingServer()

	app.CW = cw.NewCostWatcher(&app)
	go app.CW.StartWatching()

	app.LW = lw.NewLogWatcher(&app)
	go app.LW.StartLogWatcher()

	go decoy_server.Start(&app)

	app.SM = sm.NewSessionManager(&app)

	err = routes.InitRoutes(&app)
	if err != nil {
		panic(err)
	}
	logging.Info("Analyser started")
}
