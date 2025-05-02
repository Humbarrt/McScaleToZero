package conn

import (
	"bytes"
	"fmt"
	mc "mc_manager/api/mc_networking"
	"mc_manager/logging"
	"mc_manager/model"
	"net"
	"strings"
	"time"
)

type McConn struct {
	encryption  *certify
	isConnected bool
	conn        net.Conn
}

func NewMcConn(conn *net.Conn) *McConn {
	return &McConn{
		encryption:  nil,
		isConnected: true,
		conn:        *conn,
	}
}

func (c *McConn) CreateEncryption() error {
	if !c.isConnected {
		return fmt.Errorf("connection is not established")
	}
	if c.encryption != nil {
		return fmt.Errorf("encryption is already created")
	}

	encryption, err := newCert()
	if err != nil {
		logging.Error("Error creating encryption:", "error", err)
		return err
	}

	c.encryption = encryption
	return nil
}

func (c *McConn) EnableEncryption(encryptedSharedSecret []byte, encryptedVerifyToken []byte) error {
	if !c.isConnected {
		return fmt.Errorf("connection is not established")
	}
	if c.encryption == nil {
		return fmt.Errorf("no encryption created")
	}

	verifyToken, err := c.encryption.AuthDecrypt(encryptedVerifyToken)
	if err != nil {
		logging.Error("Error decrypting verify token:", "error", err)
	}
	if !bytes.Equal(c.encryption.GetVerifyToken(), verifyToken) {
		return fmt.Errorf("verifyToken is not the same as the one sent in the encryption request")
	}
	logging.Debug("VerifyToken is valid")

	sharedSecret, err := c.encryption.AuthDecrypt(encryptedSharedSecret)
	if err != nil {
		logging.Error("Error decrypting shared secret:", "error", err)
		return err
	}

	return c.encryption.EnableEncryption(sharedSecret)
}

func (c *McConn) GetEncryption() (model.McCert, error) {
	if c.encryption == nil {
		return nil, fmt.Errorf("encryption is not created")
	}
	return c.encryption, nil
}

func (c *McConn) CloseWithError(reason model.DisconnectReason, comment string, err error) {
	logging.Error(comment, "error", err)
	if !c.isConnected || c.conn == nil || strings.Contains(err.Error(), "use of closed network connection") {
		return
	}

	err = mc.WriteDisconnect(c.conn, string(reason))
	if err != nil {
		logging.Error("Error writing disconnect:", "error", err)
	}

	c.isConnected = false
	err = c.conn.Close()
	if err != nil {
		logging.Error("Error closing connection:", "error", err)
	}
}

func (c *McConn) IsConnected() bool {
	return c.isConnected
}

func (c *McConn) IsEncrypted() bool {
	return c.encryption != nil && c.encryption.IsUsed()
}

func (c *McConn) GetServerId() (string, error) {
	if c.encryption == nil {
		return "", fmt.Errorf("encryption is not created")
	}
	return c.encryption.GetServerId()
}

func (c *McConn) Read(b []byte) (n int, err error) {
	if !c.isConnected {
		return 0, fmt.Errorf("connection is not established")
	}
	if c.encryption != nil && c.encryption.IsUsed() {
		logging.Debug("Reading encrypted data")

		n, err := c.conn.Read(b)
		if err != nil {
			logging.Error("Error reading encrypted data:", "error", err)
		}
		decrypted := c.encryption.Decrypt(b[:n])
		copy(b, decrypted)
		n = len(decrypted)

		return n, err
	}
	return c.conn.Read(b)
}

func (c *McConn) Write(b []byte) (n int, err error) {
	if !c.isConnected {
		return 0, fmt.Errorf("connection is not established")
	}

	if c.encryption != nil && c.encryption.IsUsed() {
		return c.conn.Write(c.encryption.Encrypt(b))
	}

	return c.conn.Write(b)
}

func (c *McConn) Close() error {
	return c.conn.Close()
}
func (c *McConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
func (c *McConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
func (c *McConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}
func (c *McConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
func (c *McConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
