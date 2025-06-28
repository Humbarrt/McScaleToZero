package sessions

import (
	"time"

	"github.com/mileusna/useragent"
)

type UserSession struct {
	id         string
	playerId   int64
	createdAt  time.Time
	expiresAt  time.Time
	lastUsedAt time.Time
	maxAgeAt   time.Time

	userAgent *useragent.UserAgent
}

func (session *UserSession) GetId() string {
	return session.id
}

func (session *UserSession) GetPlayerId() int64 {
	return session.playerId
}

func (session *UserSession) GetCreatedAt() time.Time {
	return session.createdAt
}

func (session *UserSession) GetExpiresAt() time.Time {
	return session.expiresAt
}

func (session *UserSession) IsExpired() bool {
	return session.expiresAt.Before(time.Now())
}

func (session *UserSession) GetLastUsedAt() time.Time {
	return session.lastUsedAt
}

func (session *UserSession) ExtendExpiration() {
	newExpiration := time.Now().Add(7 * 24 * time.Hour)
	if session.maxAgeAt.Before(newExpiration) {
		session.expiresAt = session.maxAgeAt
	} else {
		session.expiresAt = newExpiration
	}
}

func (session *UserSession) UseSession() {
	session.lastUsedAt = time.Now()
}

func (session *UserSession) GetUserAgent() useragent.UserAgent {
	return *session.userAgent
}
