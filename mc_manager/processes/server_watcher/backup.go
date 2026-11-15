package server_watcher

import (
	"mc_manager/api/ssh"
	"mc_manager/logging"
	"mc_manager/model"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

func getBackupPath(dataPath string) (string, error) {
	backupPath := path.Join(dataPath, "backups")
	err := os.MkdirAll(backupPath, 0755)
	if err != nil {
		logging.Error("Error creating backup path:", "error", err)
		return "", err
	}
	return backupPath, nil
}

func BackupWorld(app *model.App) error {
	backupPath, err := getBackupPath(app.Config.GetDataPath())
	if err != nil {
		logging.Error("Error getting backup path:", "error", err)
		return err
	}

	backups, err := getBackups(backupPath)
	if err != nil {
		logging.Error("Error getting backups:", "error", err)
		return err
	}
	if len(backups) > app.Config.GetBackupsToKeep() {
		logging.Info("Deleting old backups...")
		for _, backup := range backups[app.Config.GetBackupsToKeep():] {
			err = os.Remove(path.Join(backupPath, backup.Format("2006-01-02")+".zip"))
			if err != nil {
				logging.Error("Error deleting backup:", "error", err)
				return err
			}
			logging.Info("Deleted backup:", "backup", backup.Format("2006-01-02"))
		}
	}

	backupFile := path.Join(backupPath, time.Now().UTC().Format("2006-01-02")+".zip")
	if _, err := os.Stat(backupFile); err == nil {
		logging.Info("Backup already exists, skipping backup.")
		return nil
	}

	err = downloadMcWorld(app, backupFile)
	if err != nil {
		logging.Warn("Error downloading World Backup, trying again.", "error", err)

		time.Sleep(5 * time.Second)
		err = downloadMcWorld(app, backupFile)
		if err != nil {
			logging.Error("Error downloading World Backup, giving up.", "error", err)
			return err
		}
	}

	logging.Info("World Backup downloaded successfully.")
	return nil
}

func getBackups(backupPath string) ([]time.Time, error) {
	files, err := os.ReadDir(backupPath)
	if err != nil {
		logging.Error("Error reading backup directory:", "error", err)
		return nil, err
	}

	var backups []time.Time
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".zip") {
			continue
		}

		fileName := strings.Split(file.Name(), ".")[0]
		fileTime, err := time.Parse("2006-01-02", fileName)
		if err != nil {
			logging.Error("Error parsing backup file name:", "error", err)
			continue
		}
		backups = append(backups, fileTime)
	}

	//sort backups, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].After(backups[j])
	})

	return backups, nil
}

func downloadMcWorld(app *model.App, backupFile string) error {
	remotePath := path.Join("/home", app.Config.GetMCServerSSHUser(), "mc-server/data")
	sshClient, err := ssh.NewClient(app)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	err = sshClient.ZipFolder(path.Join(remotePath, "world"), path.Join(remotePath, "world.zip"))
	if err != nil {
		logging.Warn("Error zipping world folder, trying again:", "error", err)

		time.Sleep(5 * time.Second)
		err = sshClient.ZipFolder(path.Join(remotePath, "world"), path.Join(remotePath, "world.zip"))
		if err != nil {
			logging.Error("Error zipping world folder, giving up:", "error", err)
			return err
		}
	}
	logging.Info("World folder zipped successfully.")

	err = sshClient.DownloadFile(path.Join(remotePath, "world.zip"), backupFile)
	if err != nil {
		logging.Warn("Error downloading world.zip, trying again:", "error", err)

		time.Sleep(5 * time.Second)
		err = sshClient.DownloadFile(path.Join(remotePath, "world.zip"), backupFile)
		if err != nil {
			logging.Error("Error downloading world.zip, giving up", "error", err)
			return err
		}
	}
	return nil
}
