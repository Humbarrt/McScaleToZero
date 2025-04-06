package main

import (
	"crypto/rand"
	"embed"
	"mc_analyser/api/aws"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
	"mc_analyser/web/routes"
	"time"

	_config "mc_analyser/config"
	"mc_analyser/processes/decoy_server"
	sw "mc_analyser/processes/server_watcher"

	"github.com/gin-gonic/gin"
)

//go:embed web/static/*
var fs embed.FS

func main() {
	logging.Info("Starting MC Manager")
	database, err := db.NewDatabase()
	if err != nil {
		logging.Fatal("Error creating database connection", "error", err)
	}
	config := _config.Load(database)

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
	}

	app.AWS, err = aws.NewAWSAPI(&app)
	if err != nil {
		logging.Fatal("Error creating AWS API", "error", err)
	}

	app.SW = sw.NewServerWatcher(&app)
	logging.Info("checking if online", "isOnline", app.SW.IsServerOnline())

	go app.SW.StartWatchingServer()
	go decoy_server.Start(&app)

	err = routes.InitRoutes(&app)
	if err != nil {
		panic(err)
	}
	logging.Info("Analyser started")
}

/*
TODO: Listen to server power state
TODO: Download whitelist
TODO: Download admin list
TODO: Read whitelist and add aliases
TODO: Read cloudWatch logs
TODO: Parse logs and add to database
TODO: Find last position based on 5 logs

TODO: Check "Month-to-date cost"
TODO: Check "Total forecasted cost for current month"
TODO: Record AWS Invoice
TODO: Create Splitwise invoice
TODO: Get player history since last bill. Even if player is currently online
TODO: On config, LOG to file

TODO: Set IP address of player
TODO: Set IsAdmin of player

TODO: dont send hardcoded mc status response
TODO: what if the next minecraft version comes out? - maybe check the version on boot?
*/
