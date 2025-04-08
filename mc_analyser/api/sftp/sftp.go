package sftp

import (
	"bufio"
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"runtime"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	app    *model.App
	conn   *ssh.Client
	client *sftp.Client
	file   *sftp.File
}

func NewSFTPClient(app *model.App) (*Client, error) {
	client := &Client{
		app: app,
	}
	runtime.SetFinalizer(client, func(s *Client) {
		s.Close()
	})
	return client, nil
}

func (s *Client) Close() {
	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			logging.Error("Error closing SSH connection", "error", err)
		}
	}
	if s.client != nil {
		err := s.client.Close()
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
	s.client = sftpClient
	logging.Debug("SFTP client connected", "host", ipv4, "user", s.app.Config.GetMCServerSSHUser())
	return nil
}

func (s *Client) ReadFile(remotePath string) (*bufio.Reader, error) {
	err := s.Connect()
	if err != nil {
		logging.Warn("Error connecting to SFTP server, trying again:", "error", err)
		time.Sleep(5 * time.Second)
		err = s.Connect()
		if err != nil {
			logging.Error("Error connecting to SFTP server, giving up:", "error", err)
			return nil, err
		}
	}

	remoteFile, err := s.client.Open(remotePath)
	if err != nil {
		logging.Error("Error opening remote file", "error", err)
		return nil, err
	}
	s.file = remoteFile
	return bufio.NewReader(remoteFile), nil
}
