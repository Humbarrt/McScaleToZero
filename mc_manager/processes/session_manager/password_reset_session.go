package sessions

import (
	"crypto/rsa"
	"crypto/x509"
	"time"
)

type PasswordResetSession struct {
	id              string
	playerId        int64
	chatToken       string
	isAuthenticated bool
	isApplied       bool

	createdAt time.Time
	expiresAt time.Time

	clientKey         *rsa.PublicKey
	encryptedPassword *[]byte
}

func (session *PasswordResetSession) GetId() string {
	return session.id
}

func (session *PasswordResetSession) GetPlayerId() int64 {
	return session.playerId
}

func (session *PasswordResetSession) GetClientKey() []byte {
	return x509.MarshalPKCS1PublicKey(session.clientKey)
}

func (session *PasswordResetSession) IsExpired() bool {
	return session.expiresAt.Before(time.Now())
}

func (session *PasswordResetSession) GetChatToken() string {
	return session.chatToken
}

func (session *PasswordResetSession) GetExpiration() time.Time {
	return session.expiresAt
}

func (session *PasswordResetSession) GetIsAuthenticated() bool {
	return session.isAuthenticated
}

func (session *PasswordResetSession) SetAuthenticated() {
	session.isAuthenticated = true
}

func (session *PasswordResetSession) GetIsApplied() bool {
	return session.isApplied
}
func (session *PasswordResetSession) SetApplied() {
	session.isApplied = true
}

func (session *PasswordResetSession) GetEncryptedPassword() []byte {
	if session.encryptedPassword == nil {
		return nil
	}
	return *session.encryptedPassword
}

func (session *PasswordResetSession) SetEncryptedPassword(password []byte) error {
	session.encryptedPassword = &password
	session.expiresAt = time.Now().Add(10 * time.Minute)
	return nil
}
