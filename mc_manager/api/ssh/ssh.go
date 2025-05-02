package ssh

import (
	"bufio"
	"fmt"
	"mc_manager/logging"
	model "mc_managermodel"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	app        *model.App
	conn       *ssh.Client
	sftpClient *sftp.Client
	file       *sftp.File
}

func NewClient(app *model.App) (*Client, error) {
	client := &Client{
		app: app,
	}
	err := client.Connect()
	if err != nil {
		logging.Warn("Error connecting to SFTP server, trying again:", "error", err)
		time.Sleep(5 * time.Second)
		err = client.Connect()
		if err != nil {
			logging.Error("Error connecting to SFTP server, giving up:", "error", err)
			return nil, err
		}
	}

	return client, nil
}

func (s *Client) Close() {
	if s.conn != nil {
		logging.Debug("Closing SSH connection")
		err := s.conn.Close()
		if err != nil {
			logging.Error("Error closing SSH connection", "error", err)
		}
	}
	if s.sftpClient != nil {
		logging.Debug("Closing SFTP client")
		err := s.sftpClient.Close()
		if err != nil {
			logging.Error("Error closing SFTP client", "error", err)
		}
	}
}

func (s *Client) Connect() error {
	s.Close()
	signer, err := ssh.ParsePrivateKey([]byte(s.app.Config.GetMCServerSSHKey()))
	if err != nil {
		logging.Error("Error parsing private key", "error", err)
		return err
	}

	// SSH config
	config := &ssh.ClientConfig{
		User: s.app.Config.GetMCServerSSHUser(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use known_hosts in production!
	}

	ipv4, err := s.app.SW.IPv4()
	if err != nil {
		return err
	}

	// SSH connection
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ipv4), config)
	if err != nil {
		return err
	}
	s.conn = conn

	// Create SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	s.sftpClient = sftpClient
	logging.Debug("SFTP client connected", "host", ipv4, "user", s.app.Config.GetMCServerSSHUser())
	return nil
}

func (s *Client) ZipFolder(remoteFolder, outputZip string) error {
	session, err := s.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("zip -r %s %s", outputZip, remoteFolder)
	return session.Run(cmd)
}

func (s *Client) ReadFile(remotePath string) (*bufio.Reader, error) {
	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		logging.Error("Error opening remote file", "error", err)
		return nil, err
	}
	s.file = remoteFile
	return bufio.NewReader(remoteFile), nil
}

func (s *Client) DownloadFile(remotePath, localPath string) error {
	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		logging.Error("Error opening remote file", "error", err)
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		logging.Error("Error creating local file", "error", err)
		return err
	}
	defer localFile.Close()

	_, err = remoteFile.WriteTo(localFile)
	if err != nil {
		logging.Error("Error writing to local file", "error", err)
		return err
	}

	return nil
}
