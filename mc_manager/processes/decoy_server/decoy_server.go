package decoy_server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mc "mc_manager/api/mc_networking"
	"mc_manager/api/mc_networking/conn"
	mcp "mc_manager/api/mc_networking/packet"
	"mc_manager/api/mojang"
	"mc_manager/db/queries"
	"mc_manager/logging"
	"mc_manager/model"
	"net"
	"strconv"
	"strings"
	"time"
)

func handleServerListPing(clientConn *conn.McConn, app *model.App) {
	logging.Info("Player is pinging the server", "remoteAddr", clientConn.RemoteAddr().String())

	for clientConn.IsConnected() {
		// receive packet
		packet, err := mcp.ReadPacket(clientConn)
		if err != nil {
			clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading packet:", err)
			return
		}

		if packet.PacketID == 0x00 {
			motd := []byte("")
			replaceMOTD := app.Config.GetReplaceMOTD()
			offlineMotdEscaped := app.Config.GetMCOfflineMOTD()
			offlineMotd, err := strconv.Unquote(`"` + string(offlineMotdEscaped) + `"`)
			if err != nil {
				offlineMotd = "error parsing motd"
			}
			onlineMotdEscaped := app.Config.GetMCOnlineMOTD()
			onlineMotd, err := strconv.Unquote(`"` + string(onlineMotdEscaped) + `"`)
			if err != nil {
				onlineMotd = "error parsing motd"
			}
			lastMotd := app.Config.GetLastMOTD()

			lastImage := app.Config.GetLastImage()
			protocolVersion := json.RawMessage(app.Config.GetProtocolVersion())
			maxPlayers := app.Config.GetMaxPlayers()
			statusResponse := &model.McStatusResponse{
				Version:            protocolVersion,
				Favicon:            "",
				EnforcesSecureChat: true,
				Description:        "",
				Players: struct {
					Max    int `json:"max"`
					Online int `json:"online"`
				}{
					Max:    maxPlayers,
					Online: 0,
				},
			}

			isServerReady := app.SW.IsServerReady()
			if isServerReady {
				logging.Debug("MC Server is ready, sending status response")
				ipv4, err := app.SW.IPv4()
				if err != nil {
					clientConn.CloseWithError(model.ReasonInternalServerError, "Error getting server IP:", err)
					return
				}

				status, err := mc.ReadServerStatus(ipv4, app.Config.GetMCServerPort())
				if err != nil || status == nil {
					logging.Error("Error reading server status:", "error", err)
				} else {
					statusResponse = status
					if !bytes.Equal(status.Version, protocolVersion) {
						err := app.Config.SetProtocolVersion(status.Version)
						if err != nil {
							logging.Error("Error setting protocol version:", "error", err)
						}
					}

					if lastImage == nil || (len(status.Favicon) != 0 && status.Favicon != *lastImage) {
						lastImage = &status.Favicon
						err := app.Config.SetLastImage(lastImage)
						if err != nil {
							logging.Error("Error setting last image:", "error", err)
						}
					}

					if len(status.Description) != 0 && !bytes.Equal([]byte(status.Description), lastMotd) {
						lastMotd = []byte(status.Description)
						err := app.Config.SetLastMOTD(lastMotd)
						if err != nil {
							logging.Error("Error setting last motd:", "error", err)
						}
					}

					if status.Players.Max != 0 && status.Players.Max != maxPlayers {
						maxPlayers = status.Players.Max
						err := app.Config.SetMaxPlayers(status.Players.Max)
						if err != nil {
							logging.Error("Error setting max players:", "error", err)
						}
					}

					app.EB.Broadcast(model.EB_EVENT_SERVER_INFO_UPDATE)
				}
			}

			var favicon *string
			if isServerReady {
				favicon = app.Config.GetMCOnlineImage()
			} else {
				favicon = app.Config.GetMCOfflineImage()
			}

			faviconString := ""
			if (favicon == nil || *favicon == "") && lastImage != nil {
				faviconString = string(*lastImage)
			} else if favicon != nil {
				faviconString = string(*favicon)
			}

			statusResponse.Favicon = faviconString

			switch {
			case !replaceMOTD && !isServerReady:
				motd = append(lastMotd, offlineMotd...)
			case replaceMOTD && !isServerReady:
				motd = []byte(offlineMotd)
			case !replaceMOTD && isServerReady:
				motd = append(lastMotd, onlineMotd...)
			case replaceMOTD && isServerReady:
				motd = []byte(onlineMotd)
			}
			statusResponse.Description = string(motd)
			statusResponse.Players.Max = maxPlayers

			statusResponseBytes, err := json.Marshal(statusResponse)
			if err != nil {
				clientConn.CloseWithError(model.ReasonInternalServerError, "Error marshalling status response:", err)
				return
			}

			mc.WriteStatusResponse(clientConn, string(statusResponseBytes))
		} else if packet.PacketID == 0x01 {
			mc.WritePongResponse(clientConn, packet.Data)
			clientConn.Close()
		}
	}
}

func findPlayerUUID(playerName string, app *model.App) (*model.UUID, error) {
	player, err := queries.GetPlayerByAlias(app.Db, playerName)
	uuid := &player.McUUID
	if err == nil {
		logging.Info("Player found in database", "playerName", playerName, "uuid", uuid)
		return uuid, nil
	}

	logging.Info("Player not found in database, requesting UUID from Mojang", "playerName", playerName)
	uuid, err = mojang.GetPlayerUUID(playerName)
	if err != nil {
		logging.Error("Error getting player UUID from Mojang:", "error", err)
		return &model.UUID{}, err
	}

	logging.Info("Found UUID:", "uuid", uuid)
	return uuid, nil
}

func handlePlayerLogin(clientConn *conn.McConn, app *model.App) {
	logging.Info("Player is connecting to the server", "remoteAddr", clientConn.RemoteAddr().String())

	//Wait for Login packet
	packet, err := mcp.ReadPacket(clientConn)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading packet:", err)
		return
	}
	loginStart, err := mc.ParseLoginStart(packet)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading login start:", err)
		return
	}

	// Find player UUID in db or request from Mojang
	uuid := loginStart.UUID
	if loginStart.UUID == nil {
		idound, err := findPlayerUUID(loginStart.Username, app)
		if err != nil {
			clientConn.CloseWithError(model.ReasonInternalServerError, "Error finding player UUID:", err)
			return
		}
		uuid = idound
	}

	logging.Info("Player is trying to log in", "username", loginStart.Username, "uuid", uuid)
	player, err := queries.GetPlayerByUUID(app.Db, *uuid)
	if err == sql.ErrNoRows {
		clientConn.CloseWithError(model.ReasonNotAllowed, "Error player not found in db:", err)
		return
	} else if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error querying player:", err)
		return
	}

	if !player.IsAllowed {
		err := fmt.Errorf("username: %s, UUID: %s ", loginStart.Username, loginStart.UUID.LongString())
		clientConn.CloseWithError(model.ReasonNotAllowed, "Error player is not allowed to connect:", err)
		return
	}

	// Send EncryptionRequest packet
	err = clientConn.CreateEncryption()
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error creating encryption:", err)
		return
	}

	err = mc.WriteEncryptionRequest(clientConn)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error writing encryption request:", err)
		return
	}

	// Wait for EncryptionResponse packet
	packet, err = mcp.ReadPacket(clientConn)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading encryptionResponse packet:", err)
		return
	}

	encryptionResponse, err := mc.ParseEncryptionResponse(packet)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading encryption response:", err)
		return
	}

	err = clientConn.EnableEncryption(encryptionResponse.SharedSecret.Data, encryptionResponse.VerifyToken.Data)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error enabling encryption key:", err)
		return
	}

	// Verify player session with Mojang
	playerIPV4 := clientConn.RemoteAddr().String()
	playerIPV4 = strings.Split(playerIPV4, ":")[0]
	serverId, err := clientConn.GetServerId()
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error getting server ID:", err)
		return
	}

	sessionUUID, err := mojang.GetPlayerSession(serverId, loginStart.Username, playerIPV4)
	if err != nil || sessionUUID == nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error getting player session:", err)
		return
	}

	if !sessionUUID.IsEqual(&player.McUUID) {
		err := fmt.Errorf("player UUID does not match session UUID: %s != %s", sessionUUID, player.McUUID)
		clientConn.CloseWithError(model.ReasonNotAllowed, "Error Imposter detected, kicking player:", err)
		return
	}
	logging.Info("Player session was verified by mojang")

	// Update player information in the database
	err = queries.AddPlayerAlias(app.Db, player.Id, loginStart.Username)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error adding player alias:", err)
		return
	}

	err = queries.UpdatePlayerLastIPV4(app.Db, player.Id, &playerIPV4)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error setting player IP:", err)
		return
	}

	// Send LoginSuccess packet
	loginSuccess := mc.WriteLoginSuccess(clientConn, *uuid, loginStart.Username)
	if loginSuccess != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error writing login success:", err)
		return
	}
	logging.Info("Sent LoginSuccess", "username", loginStart.Username)

	// Wait for LoginAcknowledged packet
	packet, err = mcp.ReadPacket(clientConn)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading packet:", err)
		return
	}
	loginAcknowledged, err := mc.ParseLoginAcknowledged(packet)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading login acknowledged:", err)
		return
	}
	logging.Info("Player acknowledged login", "loggedIn", loginAcknowledged, "username", loginStart.Username)

	go app.AWS.StartInstance()
	go app.SW.StartWatchingServer()

	for {
		if app.SW.IsServerReady() {
			break
		}

		logging.Debug("Server not ready, stalling player...")
		_, err = clientConn.Write([]byte{0x00})
		if err != nil {
			clientConn.CloseWithError(model.ReasonInternalServerError, "Error writing stall packet:", err)
			return
		}
		time.Sleep(1 * time.Second)
	}

	// Send Transfer packet
	ipv4, err := app.SW.IPv4()
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error getting server IP:", err)
		return
	}

	err = mc.WriteTransfer(clientConn, ipv4, app.Config.GetMCServerPort())
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error writing transfer:", err)
		return
	}
	logging.Info("Sent Transfer packet", "host", ipv4, "port", app.Config.GetMCServerPort())
}

func handleLegacyPing(clientConn net.Conn, app *model.App) {
	mc.WriteLegacyStatusResponse(clientConn, "127", "1.21.5", "Disco, Disco! Party, Party! (not started yet)", "0", "10")
}

// Handle incoming player connections
func handleConnection(clientConn *conn.McConn, app *model.App) {
	logging.Info("Player attempting to connect...", "remoteAddr", clientConn.RemoteAddr().String())
	packet, err := mcp.ReadPacket(clientConn)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error reading packet:", err)
		return
	}

	if packet.PacketID == 122 {
		logging.Info("Legacy ping packet received")
		handleLegacyPing(clientConn, app)
		clientConn.Close()
		return
	}

	handshake, err := mc.ParseHandshake(packet)
	if err != nil {
		clientConn.CloseWithError(model.ReasonInternalServerError, "Error parsing handshake:", err)
		return
	}
	logging.Info("Player handshake received", "handshake", handshake)

	if handshake.NextState == 1 {
		handleServerListPing(clientConn, app)
	} else if handshake.NextState == 2 {
		handlePlayerLogin(clientConn, app)
	} else {
		logging.Error("Unknown handshake state", "state", handshake.NextState)
		clientConn.Close()
	}
}

func Start(app *model.App) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.GetTransferServerPort()))
	if err != nil {
		logging.Fatal("Failed to start TCP Proxy:", "error", err)
	}
	defer listener.Close()

	logging.Info("TCP Proxy listening on port " + fmt.Sprintf("%d", app.Config.GetTransferServerPort()) + "...")

	for {
		logging.Debug("Waiting for incoming connections...")
		clientConn, err := listener.Accept()
		if err != nil {
			logging.Error("Connection failed:", "error", err)
			continue
		}

		mcConn := conn.NewMcConn(&clientConn)
		go handleConnection(mcConn, app)
	}
}
