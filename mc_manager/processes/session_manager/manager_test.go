package sessions_test

import (
	"fmt"
	"mc_manager/logging"
	sessions "mc_manager/processes/session_manager"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetSession(t *testing.T) {
	sm := sessions.NewSessionManager(nil)
	sm.CleanupPasswordResetSessions()

	playerId := int64(1)
	chatTokenRegex := regexp.MustCompile(`(?i)mcbillsplitter reset-password ([a-zA-Z0-9]{16})`)

	session, err := sm.CreatePasswordResetSession(playerId)
	assert.NoError(t, err)

	sessionId := session.GetId()
	chatToken := session.GetChatToken()
	clientKey := session.GetClientKey()
	assert.NotEmpty(t, sessionId)
	assert.NotEmpty(t, chatToken)
	assert.NotEmpty(t, clientKey)
	assert.True(t, chatTokenRegex.MatchString(chatToken))
	assert.False(t, session.IsExpired())

	session, err = sm.GetPasswordResetSessionById(playerId, sessionId)
	assert.NoError(t, err)
	assert.Equal(t, sessionId, session.GetId())
	assert.Equal(t, chatToken, session.GetChatToken())
	assert.Equal(t, clientKey, session.GetClientKey())

	logging.Debug("Session ID", "sessionId", session.GetId())
	logging.Debug("Chat Token", "chatToken", session.GetChatToken())
	logging.Debug("Client Key", "clientKey", fmt.Sprintf("%x", session.GetClientKey()[:]))
}
