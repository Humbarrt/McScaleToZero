package graphql

import (
	"bytes"
	"mc_analyser/logging"
	"os"
	"os/exec"
	"strings"
)

type SFTPClient struct {
	FLY_API_TOKEN string
}

func NewSFTPClient(FLY_API_TOKEN string) *SFTPClient {
	return &SFTPClient{
		FLY_API_TOKEN: FLY_API_TOKEN,
	}
}

func (c *SFTPClient) Connect() error {
	os.MkdirAll("./keys", 0755)

	//execute command fly ssh issue personal ./keys/private -o personal
	cmd := exec.Command("fly", "auth", "whoami", "-t", strings.Split(c.FLY_API_TOKEN, " ")[1])
	cmd.Env = append(os.Environ(), "FLY_API_TOKEN="+c.FLY_API_TOKEN)
	logging.Info("Executing command", "command", cmd.String())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logging.Error("Error executing command", "error", err)
		logging.Error("Stderr Output", "stderr", stderr.String())
		return err
	}
	logging.Info("Command executed successfully", "stdout", stdout.String())

	//check if private key exists
	_, err = os.Stat("./keys/private")
	if os.IsNotExist(err) {
		logging.Error("Private key does not exist", "error", err)
		return err
	}

	return nil
}
