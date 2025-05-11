package model

import (
	"encoding/json"
	"net"
	"time"
)

type DisconnectReason string

const (
	ReasonInternalServerError DisconnectReason = "Internal Server Error"
	ReasonNotAllowed          DisconnectReason = "Player Not Allowed"
)

type McConn interface {
	CreateEncryption() error
	EnableEncryption(encryptedSharedSecret []byte, encryptedVerifyToken []byte) error
	GetEncryption() (McCert, error)
	CloseWithError(reason DisconnectReason, comment string, err error)
	IsConnected() bool
	IsEncrypted() bool
	GetServerId() (string, error)
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
}

type McCert interface {
	IsUsed() bool
	GetPublicKey() []byte
	GetPrivateKey() []byte
	GetVerifyToken() []byte
	GetServerId() (string, error)
	AuthEncrypt(data []byte) ([]byte, error)
	AuthDecrypt(encrypted []byte) ([]byte, error)
	EnableEncryption(sharedSecret []byte) error
	Encrypt(data []byte) (output []byte)
	Decrypt(data []byte) (output []byte)
}

type McStatusResponse struct {
	Version            json.RawMessage `json:"version"`
	Favicon            string          `json:"favicon"`
	EnforcesSecureChat bool            `json:"enforcesSecureChat"`
	Description        string          `json:"description"`
	Players            struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`
}
