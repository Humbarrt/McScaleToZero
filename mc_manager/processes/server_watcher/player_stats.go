package server_watcher

import (
	"archive/zip"
	"fmt"
	"mc_manager/logging"
	"mc_manager/model"
	"os"
	"path"
	"time"
)

func UpdatePlayerStats(app *model.App) error {
	backupPath, err := getBackupPath(app.Config.GetDataPath())
	if err != nil {
		logging.Error("Error getting backup path:", "error", err)
		return err
	}

	backupFile := path.Join(backupPath, time.Now().UTC().Format("2006-01-02")+".zip")
	if _, err := os.Stat(backupFile); err != nil {
		logging.Error("No backup found for today")
		return fmt.Errorf("Error no backup found for today")
	}

	read, err := zip.OpenReader(backupFile)
	if err != nil {
		logging.Error("Error reading world zip", "error", err)
		return err
	}
	defer read.Close()

	for _, f := range read.File {
		logging.Info("Zip File", "name", f.Name)
	}

	return nil
}
