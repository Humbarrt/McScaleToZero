package main

import (
	"embed"
	"mc_analyser/api/aws"
	"mc_analyser/db"
	"mc_analyser/logging"
	"mc_analyser/model"
	"mc_analyser/web/routes"
	"time"

	_config "mc_analyser/config"
	cw "mc_analyser/processes/cost_watcher"
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
	/* _, err = rand.Read(secret)
	if err != nil {
		logging.Fatal("Error creating JWTSecret", "error", err)
	} */

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
	}

	app.AWS, err = aws.NewAWSAPI(&app)
	if err != nil {
		logging.Fatal("Error creating AWS API", "error", err)
	}

	app.SW = sw.NewServerWatcher(&app)
	go app.SW.StartWatchingServer()

	app.CW = cw.NewCostWatcher(&app)
	go app.CW.StartWatching()

	go decoy_server.Start(&app)

	err = routes.InitRoutes(&app)
	if err != nil {
		panic(err)
	}
	logging.Info("Analyser started")
}

/*
TODO: get cost explorer by usage type

TODO: dynamic decoy response, based on the last version we saw.
TODO: dynamic image/motd/max-player-count replacement
TODO: show different text on online/started/offline

TODO: Read cloudWatch logs
TODO: Parse logs and add to database
TODO: Find last position based on 5 logs

TODO: Check "Month-to-date cost"
TODO: Check "Total forecasted cost for current month"
TODO: Record AWS Invoice
TODO: Create Splitwise invoice
TODO: Get player history since last bill. Even if player is currently online
TODO: On config, LOG to file

TODO: dont send hardcoded mc status response
TODO: what if the next minecraft version comes out? - maybe check the version on boot?
*/
