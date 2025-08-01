package model

import (
	"time"

	"github.com/mileusna/useragent"
)

type PasswordResetSession interface {
	GetId() string
	GetPlayerId() int64
	GetClientKey() []byte
	GetChatToken() string
	GetEncryptedPassword() []byte
	SetEncryptedPassword(password []byte) error
	GetExpiration() time.Time
	IsExpired() bool
	GetIsAuthenticated() bool
	SetAuthenticated()
	GetIsApplied() bool
	SetApplied()
}

type UserSession interface {
	GetId() string
	GetPlayerId() int64
	GetCreatedAt() time.Time
	GetExpiresAt() time.Time
	GetLastUsedAt() time.Time
	IsExpired() bool
	GetUserAgent() useragent.UserAgent
	UseSession()
	ExtendExpiration()
}

type SessionManager interface {
	CleanupUserSessions()
	CreateUserSession(playerId int64, ua *useragent.UserAgent) (UserSession, error)
	GetUserSessionById(playerId int64, sessionId string) (UserSession, error)
	ExtendSession(playerId int64, sessionId string) (UserSession, error)
	GetPlayerSessions(playerId int64) ([]UserSession, error)
	CleanupPasswordResetSessions()
	GetPasswordResetSessionById(playerId int64, sessionId string) (PasswordResetSession, error)
	AuthenticatePasswordResetSession(playerId int64, chatToken string) error
	CreatePasswordResetSession(playerId int64) (PasswordResetSession, error)
	SetResetPassword(playerId int64, sessionId string, password []byte) error
	RemoveUserSessions(playerId int64) error
	RemoveUserSession(playerId int64, sessionId string) error
	ConsumePasswordResetSession(playerId int64, sessionId string) (PasswordResetSession, error)
}
