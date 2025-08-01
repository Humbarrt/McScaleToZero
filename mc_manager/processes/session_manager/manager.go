package sessions

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
)

/*
PasswordResetSession works by authenticating via text in chat
loginSession works by authenticating via password
userSessions are the sessions that are created when a user logs in
*/

func NewSessionManager(app *model.App) *SessionManager {
	return &SessionManager{
		app:                    app,
		_passwordResetSessions: make(map[int64]*PasswordResetSession),
		_userSessions:          make(map[int64]map[string]*UserSession),
	}
}

type SessionManager struct {
	app *model.App
	//protected
	mutex                  sync.Mutex
	_passwordResetSessions map[int64]*PasswordResetSession
	_userSessions          map[int64]map[string]*UserSession
}

func (sm *SessionManager) CleanupUserSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for _, playeressions := range sm._userSessions {
		for sessionId, session := range playeressions {
			if session.IsExpired() {
				delete(playeressions, sessionId)
			}
		}
	}
}

func (sm *SessionManager) CreateUserSession(playerId int64, ua *useragent.UserAgent) (model.UserSession, error) {
	sm.CleanupUserSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionId := uuid.New().String()
	session := &UserSession{
		id:        sessionId,
		playerId:  playerId,
		createdAt: time.Now(),
		expiresAt: time.Now().Add(7 * 24 * time.Hour),
		maxAgeAt:  time.Now().Add(60 * 24 * time.Hour),
		userAgent: ua,
	}

	if _, ok := sm._userSessions[playerId]; !ok {
		sm._userSessions[playerId] = make(map[string]*UserSession)
	}

	sm._userSessions[playerId][sessionId] = session
	copy := *session
	return &copy, nil
}

func (sm *SessionManager) GetUserSessionById(playerId int64, sessionId string) (model.UserSession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	playeressions, ok := sm._userSessions[playerId]
	if !ok {
		return &UserSession{}, fmt.Errorf("session not found")
	}

	session, ok := playeressions[sessionId]
	if !ok {
		return &UserSession{}, fmt.Errorf("session not found")
	}

	if session.IsExpired() {
		delete(playeressions, sessionId)
		return &UserSession{}, fmt.Errorf("session expired")
	}

	copy := *session
	return &copy, nil
}

func (sm *SessionManager) ExtendSession(playerId int64, sessionId string) (model.UserSession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	playeressions, ok := sm._userSessions[playerId]
	if !ok {
		return &UserSession{}, fmt.Errorf("session not found")
	}

	session, ok := playeressions[sessionId]
	if !ok {
		return &UserSession{}, fmt.Errorf("session not found")
	}

	if session.IsExpired() {
		delete(playeressions, sessionId)
		return &UserSession{}, fmt.Errorf("session expired")
	}

	session.ExtendExpiration()
	session.UseSession()

	copy := *session
	return &copy, nil
}

func (sm *SessionManager) GetPlayerSessions(playerId int64) ([]model.UserSession, error) {
	sm.CleanupUserSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	playeressions, ok := sm._userSessions[playerId]
	if !ok {
		return nil, fmt.Errorf("no sessions found for player %d", playerId)
	}
	var sessions []model.UserSession
	for _, session := range playeressions {
		if session.IsExpired() {
			delete(playeressions, session.GetId())
			continue
		}
		copy := *session
		sessions = append(sessions, &copy)
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no valid sessions found for player %d", playerId)
	}

	return sessions, nil
}

func (sm *SessionManager) CleanupPasswordResetSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for playerId, session := range sm._passwordResetSessions {
		if session.IsExpired() {
			delete(sm._passwordResetSessions, playerId)
		}
	}
}

func (sm *SessionManager) GetPasswordResetSessionById(playerId int64, sessionId string) (model.PasswordResetSession, error) {
	sm.CleanupPasswordResetSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm._passwordResetSessions[playerId]
	if !ok {
		return &PasswordResetSession{}, fmt.Errorf("session not found")
	}

	if session.GetId() != sessionId {
		return &PasswordResetSession{}, fmt.Errorf("session id does not match")
	}

	copy := *session
	return &copy, nil
}

func (sm *SessionManager) applyPasswordResetSession(session model.PasswordResetSession) error {
	if !session.GetIsAuthenticated() || session.GetIsApplied() {
		return fmt.Errorf("session not authenticated")
	}

	playerId := session.GetPlayerId()
	encryptedPassword := session.GetEncryptedPassword()
	if encryptedPassword == nil {
		return fmt.Errorf("password not set")
	}

	clientKey := session.GetClientKey()
	if clientKey == nil {
		return fmt.Errorf("client key not set")
	}

	err := queries.UpdatePlayerPassword(sm.app.Db, playerId, encryptedPassword, clientKey)
	if err != nil {
		logging.Warn("ApplyPasswordResetSession error updating player password, trying again", "error", err)
		err = queries.UpdatePlayerPassword(sm.app.Db, playerId, encryptedPassword, clientKey)
		if err != nil {
			logging.Error("ApplyPasswordResetSession error updating player password", "error", err)
			return fmt.Errorf("error updating player password: %w", err)
		}
	}

	delete(sm._userSessions, playerId)
	session.SetApplied()
	return nil
}

func (sm *SessionManager) ConsumePasswordResetSession(playerId int64, sessionId string) (model.PasswordResetSession, error) {
	sm.CleanupPasswordResetSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm._passwordResetSessions[playerId]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if session.GetId() != sessionId {
		return nil, fmt.Errorf("session id does not match")
	}

	if !session.GetIsAuthenticated() {
		return nil, fmt.Errorf("session not authenticated")
	}

	if !session.GetIsApplied() {
		sm.applyPasswordResetSession(session)
	}

	delete(sm._passwordResetSessions, playerId)
	return session, nil
}

func (sm *SessionManager) AuthenticatePasswordResetSession(playerId int64, chatToken string) error {
	sm.CleanupPasswordResetSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm._passwordResetSessions[playerId]
	if !ok {
		return fmt.Errorf("session not found")
	}

	if session.GetChatToken() != chatToken {
		return fmt.Errorf("chat token does not match chatToken: %s sesionChatToken: %s", chatToken, session.GetChatToken())
	}

	session.SetAuthenticated()
	return sm.applyPasswordResetSession(session)
}

func (sm *SessionManager) SetResetPassword(playerId int64, sessionId string, password []byte) error {
	sm.CleanupPasswordResetSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm._passwordResetSessions[playerId]
	if !ok {
		return fmt.Errorf("session not found")
	}

	if session.GetId() != sessionId {
		return fmt.Errorf("session id does not match")
	}

	if err := session.SetEncryptedPassword(password); err != nil {
		return err
	}
	return nil
}

func (sm *SessionManager) CreatePasswordResetSession(playerId int64) (model.PasswordResetSession, error) {
	sm.CleanupPasswordResetSessions()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm._passwordResetSessions, playerId)

	sessionId := uuid.New().String()
	randomToken, err := randomString(16)
	if err != nil {
		return &PasswordResetSession{}, fmt.Errorf("error generating random string: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return &PasswordResetSession{}, err
	}

	chatLoginSession := &PasswordResetSession{
		id:        sessionId,
		playerId:  playerId,
		clientKey: &privateKey.PublicKey,
		chatToken: fmt.Sprintf("mcbillsplitter reset-password %s", randomToken),
		createdAt: time.Now(),
		expiresAt: time.Now().Add(10 * time.Minute),
	}

	sm._passwordResetSessions[playerId] = chatLoginSession
	copy := *chatLoginSession
	return &copy, nil
}

func (sm *SessionManager) RemoveUserSessions(playerId int64) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm._userSessions, playerId)
	return nil
}

func (sm *SessionManager) RemoveUserSession(playerId int64, sessionId string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	playeressions, ok := sm._userSessions[playerId]
	if !ok {
		return fmt.Errorf("session not found")
	}

	if _, ok := playeressions[sessionId]; !ok {
		return fmt.Errorf("session not found")
	}

	delete(playeressions, sessionId)
	if len(playeressions) == 0 {
		delete(sm._userSessions, playerId)
	}
	return nil
}

func randomString(length int) (string, error) {
	b := make([]byte, length+2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b)[2 : length+2], nil
}
