package log_watcher

import (
	"database/sql"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"regexp"
	"strings"
	"sync"
	"time"
)

type logWatcher struct {
	App     *model.App
	retries int
	lastLog *model.Log
	limit   int

	reJoindedGame *regexp.Regexp
	reLeftGame    *regexp.Regexp
	reIsMessage   *regexp.Regexp
	reIsChatToken *regexp.Regexp

	//protected
	mutex         sync.Mutex
	_try          int
	_isWatching   bool
	_isCatchingUp bool
}

func NewLogWatcher(app *model.App) *logWatcher {
	return &logWatcher{
		App:           app,
		retries:       5,
		limit:         100,
		_isCatchingUp: true,
		lastLog:       nil,
		reJoindedGame: regexp.MustCompile(`^\[\d{2}:\d{2}:\d{2}\] \[Server thread\/INFO\]: ([A-Za-z0-9_]{3,16}) joined the game$`),
		reLeftGame:    regexp.MustCompile(`^\[\d{2}:\d{2}:\d{2}\] \[Server thread\/INFO\]: ([A-Za-z0-9_]{3,16}) left the game$`),
		reIsMessage:   regexp.MustCompile(`^\[\d{2}:\d{2}:\d{2}\] \[Server thread\/INFO\]: <([A-Za-z0-9_]{3,16})> (.*)$`),
		reIsChatToken: regexp.MustCompile(`(?i)mcbillsplitter reset-password ([a-zA-Z0-9]{16})`),
	}
}

func (lw *logWatcher) parseLog(log *model.Log) {
	if log == nil {
		return
	}

	var playerName *string
	var hasJoinedGame *bool
	var chatToken *string
	if matches := lw.reJoindedGame.FindStringSubmatch(log.Message); matches != nil {
		playerName = &matches[1]
		_hasJoined := true
		hasJoinedGame = &_hasJoined
		lw.App.EB.Broadcast(model.EB_EVENT_PLAYER_JOINED)
	} else if matches := lw.reLeftGame.FindStringSubmatch(log.Message); matches != nil {
		playerName = &matches[1]
		_hasJoined := false
		hasJoinedGame = &_hasJoined
		lw.App.EB.Broadcast(model.EB_EVENT_PLAYER_LEFT)
	} else if matches := lw.reIsMessage.FindStringSubmatch(log.Message); matches != nil {
		playerName = &matches[1]
		playerMessage := matches[2]
		if !strings.HasPrefix(playerMessage, "mcbillsplitter") {
			return
		}
		chatTokenMatches := lw.reIsChatToken.FindStringSubmatch(playerMessage)
		if chatTokenMatches == nil {
			return
		}

		chatToken = &playerMessage
	}

	if playerName == nil {
		return
	}

	player, err := queries.GetPlayerByAlias(lw.App.Db, *playerName)
	if err != nil {
		logging.Warn("LogWatcher: Error getting player id, trying again", "error", err)
		player, err = queries.GetPlayerByAlias(lw.App.Db, *playerName)
		if err != nil {
			logging.Error("LogWatcher: Error getting player id, last Try failed", "error", err)
			return
		}
	}

	if hasJoinedGame != nil {
		err := queries.AddPlayerEvent(lw.App.Db, log.Timestamp, player.Id, *hasJoinedGame)
		if err != nil {
			logging.Warn("LogWatcher: Error adding player event, trying again", "error", err)
			err = queries.AddPlayerEvent(lw.App.Db, log.Timestamp, player.Id, *hasJoinedGame)
			if err != nil {
				logging.Error("LogWatcher: Error adding player event, last Try failed", "error", err)
			}
		}

		logging.Info("LogWatcher: Player event", "player", playerName, "hasJoinedGame", *hasJoinedGame)
	} else if chatToken != nil {
		logging.Info("LogWatcher: Player authenticating passwordResetSession", "player", playerName, "chatToken", *chatToken)
		err := lw.App.SM.AuthenticatePasswordResetSession(player.Id, *chatToken)
		if err != nil {
			logging.Warn("LogWatcher: Error authenticating password reset session", "error", err)
		}

	}
}

func (lw *logWatcher) catchupLogs() {
	for {
		startTime := model.NewYearMonthFromTime(time.Now().UTC()).AddMonths(-1).StartTime()
		if lw.lastLog != nil {
			startTime = lw.lastLog.Timestamp
		}

		logs, err := lw.App.AWS.GetLogsSince(startTime, lw.limit)
		if err != nil {
			logging.Error("LogWatcher: Error getting logs", "error", err)
			break
		}
		logging.Info("LogWatcher: Got logs", "count", len(logs))

		hasFoundLog := false
		for _, log := range logs {
			if log.Timestamp.Before(startTime) || !hasFoundLog && log.Timestamp.Equal(startTime) {
				if lw.lastLog != nil && lw.lastLog.Timestamp.Equal(log.Timestamp) && lw.lastLog.Message == log.Message {
					logging.Info("LogWatcher found lastLog, skipping", "log", log)
					hasFoundLog = true
				}
				continue
			}
			lw.parseLog(log)
		}

		if len(logs) > 0 {
			lw.lastLog = logs[len(logs)-1]
			err = queries.SetLastLog(lw.App.Db, lw.lastLog)
			if err != nil {
				logging.Warn("LogWatcher: Error setting last log, trying again", "error", err)
				err = queries.SetLastLog(lw.App.Db, lw.lastLog)
				if err != nil {
					logging.Error("LogWatcher: Error setting last log, last Try failed", "error", err)
				}
			}
		}

		if len(logs) < lw.limit {
			logging.Info("LogWatcher: No new logs")
			break
		}
	}

	isOnline := lw.App.SW.IsServerOnline()
	lw.mutex.Lock()
	lw._isCatchingUp = false
	if !isOnline {
		logging.Info("LogWatcher: Server is offline", "retries left", lw.retries-lw._try)
		lw._try = lw._try + 1
	} else {
		lw._try = 0
	}
	lw.mutex.Unlock()
	time.Sleep(10 * time.Second)
}

func (lw *logWatcher) IsCatchingUp() bool {
	lw.mutex.Lock()
	defer lw.mutex.Unlock()
	return lw._isCatchingUp
}

func (lw *logWatcher) StartLogWatcher() {
	logging.Info("Trying to start LogWatcher")
	lw.mutex.Lock()
	lw._try = 0
	if lw._isWatching {
		lw.mutex.Unlock()
		logging.Error("LogWatcher is already running")
		return
	}
	lw._isWatching = true
	lw.mutex.Unlock()

	defer func() {
		lw.mutex.Lock()
		lw._isWatching = false
		lw.mutex.Unlock()
		logging.Info("LogWatcher: Stopped watching logs")
	}()

	//get last log
	lastLog, err := queries.GetLastLog(lw.App.Db)
	if err != nil && err != sql.ErrNoRows {
		logging.Warn("LogWatcher: Error getting last log, trying again", "error", err)
		lastLog, err = queries.GetLastLog(lw.App.Db)
		if err != nil && err != sql.ErrNoRows {
			logging.Error("LogWatcher: Error getting last log, last Try failed", "error", err)
			return
		}
	}
	lw.lastLog = lastLog

	for {
		lw.mutex.Lock()
		if lw._try > lw.retries {
			lw._isWatching = false
			lw.mutex.Unlock()
			break
		}
		lw.mutex.Unlock()
		lw.catchupLogs()
	}

	isOnline := lw.App.SW.IsServerOnline()
	if isOnline {
		logging.Warn("LogWatcher: Server is online, this should not happen")
		return
	}

	events, err := queries.GetLastestPlayersEvents(lw.App.Db)
	if err != nil {
		logging.Warn("LogWatcher: Error getting last player events, trying again", "error", err)
		events, err = queries.GetLastestPlayersEvents(lw.App.Db)
		if err != nil {
			logging.Error("LogWatcher: Error getting all player events", "error", err)
			return
		}
	}

	for _, event := range events {
		if !event.IsOnline {
			continue
		}
		err = queries.AddPlayerEvent(lw.App.Db, time.Now().UTC(), event.PlayerId, false)
		if err != nil {
			logging.Error("LogWatcher: Error adding player event", "error", err)
			continue
		}
	}

	lw.App.EB.Broadcast(model.EB_EVENT_PLAYER_LEFT)
	logging.Info("LogWatcher: Server stopped, all players set to offline")
}
