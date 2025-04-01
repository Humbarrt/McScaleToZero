package mc_decoy_server

import (
	"fmt"
	mc "mc_analyser/api/mcnetworking"
	"mc_analyser/logging"
	"mc_analyser/model"
	"net"
)

// Handle incoming player connections
func handleConnection(clientConn net.Conn, app *model.App) {
	logging.Info("Player attempting to connect...", "remoteAddr", clientConn.RemoteAddr().String())
	handshake, err := mc.ReadHandshake(clientConn)
	if err != nil {
		logging.Error("Error reading handshake:", "error", err)
		clientConn.Close()
		return
	}
	logging.Info("Player handshake received", "handshake", handshake)
	
	if hand
	// TODO: Add state 1
	else if handshake.NextState == 2 {
		logging.Info("Player is connecting to the server", "remoteAddr", clientConn.RemoteAddr().String())

		//Wait for Login packet
		loginStart, err := mc.ReadLoginStart(clientConn)
		if err != nil {
			logging.Error("Error reading login start:", "error", err)
			clientConn.Close()
			return
		}
		logging.Info("Player is logging in", "username", loginStart.Username, "uuid", loginStart.UUID)
		if loginStart.UUID == nil {
			//TODO: request the uuid from mojang
			logging.Error("UUID is nil, requesting from mojang")
		}

		// TODO: Return if the player is not in the whitelist

		uuid := *loginStart.UUID
		// Send LoginSuccess packet
		loginSuccess := mc.WriteLoginSuccess(clientConn, uuid, loginStart.Username)
		if loginSuccess != nil {
			logging.Error("Error writing login success:", "error", loginSuccess)
			clientConn.Close()
			return
		}
		logging.Info("Sent LoginSuccess", "username", loginStart.Username)

		// Wait for LoginAcknowledged packet
		loginAcknowledged, err := mc.ReadLoginAcknowledged(clientConn)
		if err != nil {
			logging.Error("Error reading login acknowledged:", "error", loginAcknowledged)
			clientConn.Close()
			return
		}
		logging.Info("Player acknowledged login", "loggedIn", loginAcknowledged, "username", loginStart.Username)

		//TOOD: Stall Player with one /x00 packet every second.

		// Send Transfer packet
		err = mc.WriteTransfer(clientConn, "org.playskyward.gg", 25565)
		if err != nil {
			logging.Error("Error writing transfer:", "error", err)
			clientConn.Close()
			return
		}
		logging.Info("Sent Transfer packet", "host", "org.playskyward.gg", "port", 25565)
	}
}
func Start(app *model.App) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.MC_PORT))
	if err != nil {
		logging.Fatal("Failed to start TCP Proxy:", "error", err)
	}
	defer listener.Close()

	logging.Info("TCP Proxy listening on port " + fmt.Sprintf("%d", app.Config.MC_PORT) + "...")

	for {
		logging.Debug("Waiting for incoming connections...")
		clientConn, err := listener.Accept()
		if err != nil {
			logging.Error("Connection failed:", "error", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(clientConn, app)
	}
}
