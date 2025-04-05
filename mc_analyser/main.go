package main

import (
	"embed"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
	"time"

	_config "mc_analyser/config"
	mpc "mc_analyser/processes/mc_decoy_server"

	"github.com/gin-gonic/gin"
)

//go:embed web/static/*
var fs embed.FS

func main() {
	config := _config.Load()
	database, err := db.NewDatabase()
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	app := model.App{
		Db:      database,
		Config:  config,
		Router:  router,
		FS:      fs,
		Timeout: 10 * time.Second,
	}

	app.SetMCServerReady(false)
	logging.Info("Starting mc_analyser")
	//decoyServer
	mpc.Start(&app)

	//go sw.ServerWatcher(&app)

	/* err = routes.InitRoutes(&app)
	if err != nil {
		panic(err)
	} */
	logging.Info("Analyser started")
}

/*
TODO: Listen to server power state
TODO: Download log files
TODO: Download whitelist
TODO: Read whitelist and add aliases
TODO: Read log files
TODO: Read Prometheus logs
TODO: Parse logs and add to database
TODO: Find last position based on 5 logs
TODO: Listen to Credit Balance
TODO: Record Credit Balance on change
TODO: Create bill if credit balance has gone down
TODO: Get player history since last bill. Even if player is currently online
TODO: is using utc??????
*/
