package mc_decoy_server

import (
	"fmt"
	mc "mc_analyser/api/mcnetworking"
	mcp "mc_analyser/api/mcnetworking/packet"
	"mc_analyser/api/mojang"
	"mc_analyser/db/queries"
	"mc_analyser/logging"
	"mc_analyser/model"
	"net"
	"time"
)

var defaultResponse string = `
{
    "version": 
    {
        "name": "1.21.5",
        "protocol": 770
    },
    "favicon": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAwnpUWHRSYXcgcHJvZmlsZSB0eXBlIGV4aWYAAHjabVDbDcMwCPxnio5gA3ZgHKdJpW7Q8YsNieKmJ5nXoTMA++f9gkcHZgYui1StNRlYWbFZIMnRhs2Jhx3goCyf6nASaCUyT55Kjf6jnk8Bd82ichGSZxDrTGj8gPIjhO6oT9TjLYQ0hAidyCHQfK1UVZbrCuueZog/6IZlHvuWL3a9rdg/hLhTpmSWqPoA1F8BahbwsHYOa1KLiWRUjlXtIP/udAC+39dZFo5puzIAAAGFaUNDUElDQyBwcm9maWxlAAB4nH2RPUjDUBSFT9OKIhURO4goZKiCYBcVcSxVLIKF0lZo1cHkpX/QpCFJcXEUXAsO/ixWHVycdXVwFQTBHxB3wUnRRUq8Lym0iPHC432cd8/hvfsAoVFhqhmIAqpmGal4TMzmVsXuVwTgwwBGMSExU0+kFzPwrK976qW6i/As774/q0/JmwzwicRRphsW8Qbx7Kalc94nDrGSpBCfE08adEHiR67LLr9xLjos8MyQkUnNE4eIxWIHyx3MSoZKPEMcVlSN8oWsywrnLc5qpcZa9+QvDOa1lTTXaY0gjiUkkIQIGTWUUYGFCO0aKSZSdB7z8A87/iS5ZHKVwcixgCpUSI4f/A9+z9YsTE+5ScEY0PVi2x9jQPcu0Kzb9vexbTdPAP8zcKW1/dUGMPdJer2thY+A/m3g4rqtyXvA5Q4w9KRLhuRIflpCoQC8n9E35YDBW6B3zZ1b6xynD0CGZrV8AxwcAuNFyl73eHdP59z+7WnN7wel03K7IuwYOwAADXZpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+Cjx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IlhNUCBDb3JlIDQuNC4wLUV4aXYyIj4KIDxyZGY6UkRGIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyI+CiAgPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIKICAgIHhtbG5zOnhtcE1NPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvbW0vIgogICAgeG1sbnM6c3RFdnQ9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZUV2ZW50IyIKICAgIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIKICAgIHhtbG5zOkdJTVA9Imh0dHA6Ly93d3cuZ2ltcC5vcmcveG1wLyIKICAgIHhtbG5zOnRpZmY9Imh0dHA6Ly9ucy5hZG9iZS5jb20vdGlmZi8xLjAvIgogICAgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIgogICB4bXBNTTpEb2N1bWVudElEPSJnaW1wOmRvY2lkOmdpbXA6NWE0NjgyNjktYzgzMy00MTkwLTkzN2QtNDVhYTBjMzQ5MzIzIgogICB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOjI4NzM3MThiLTAwMmQtNDkwMS1hNTlhLWQ5MjkzZjgxZjRlZCIKICAgeG1wTU06T3JpZ2luYWxEb2N1bWVudElEPSJ4bXAuZGlkOmEwNWFmZjQxLTA5NzEtNGRhYi04ODY4LTZhOTg4OWUyY2M4MCIKICAgZGM6Rm9ybWF0PSJpbWFnZS9wbmciCiAgIEdJTVA6QVBJPSIyLjAiCiAgIEdJTVA6UGxhdGZvcm09IldpbmRvd3MiCiAgIEdJTVA6VGltZVN0YW1wPSIxNzQzNTI5MDg3NTQ1ODQxIgogICBHSU1QOlZlcnNpb249IjIuMTAuMzYiCiAgIHRpZmY6T3JpZW50YXRpb249IjEiCiAgIHhtcDpDcmVhdG9yVG9vbD0iR0lNUCAyLjEwIgogICB4bXA6TWV0YWRhdGFEYXRlPSIyMDI1OjA0OjAxVDE5OjM4OjA1KzAyOjAwIgogICB4bXA6TW9kaWZ5RGF0ZT0iMjAyNTowNDowMVQxOTozODowNSswMjowMCI+CiAgIDx4bXBNTTpIaXN0b3J5PgogICAgPHJkZjpTZXE+CiAgICAgPHJkZjpsaQogICAgICBzdEV2dDphY3Rpb249InNhdmVkIgogICAgICBzdEV2dDpjaGFuZ2VkPSIvIgogICAgICBzdEV2dDppbnN0YW5jZUlEPSJ4bXAuaWlkOjg4YWU2OGM5LTc3MDEtNGRkMy05NzU4LWMzMjgyZmZjNzgwNCIKICAgICAgc3RFdnQ6c29mdHdhcmVBZ2VudD0iR2ltcCAyLjEwIChXaW5kb3dzKSIKICAgICAgc3RFdnQ6d2hlbj0iMjAyNS0wNC0wMVQxOTozODowNyIvPgogICAgPC9yZGY6U2VxPgogICA8L3htcE1NOkhpc3Rvcnk+CiAgPC9yZGY6RGVzY3JpcHRpb24+CiA8L3JkZjpSREY+CjwveDp4bXBtZXRhPgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgCjw/eHBhY2tldCBlbmQ9InciPz7xwG+8AAAABmJLR0QA/wD/AP+gvaeTAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAB3RJTUUH6QQBESYH8P3+jwAAGCRJREFUeNp1m9+SJcdx3n+ZVd3nzOwf7C5AgBRBkbIsWYqQSEVYF760X8D3flG/gy8d4QtTjIBMUiQIEIsFsLszc7qrMnWR2dV9BjAiBnvmzDndVVmZX375Zbb85G//uzvX//n3XuRLv/7U4y/69n//4e/ESwOX7Qrf+2C89OsFOCAgAiKCe7yxfV/Y/hY/KoJovlalqMbvCAjxfXfcjSqqjzYGn/7oln/78i4+nO/J92ziyCNjybYp2X733KuAey5Zry/m3zfgq2czX79dYpNHW3hcz/3xSvKSaSBTKAi1KrUWRI6ntG/e3KkiishhOy784asLqoojiF+d1dV/z58Uvn3fdgv9/xaGg8RmZNtmvEXeAHcfdnnzroOUuKx4eszBQ+ToJeze5OEdKoVaKx+9vOHd/UJRZ20dz42PH3OqbDc6nI7guGta1fnxq1s+f30ff/N9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",
    "enforcesSecureChat": true,
    "description": "Disco, Disco! Party, Party! (not started yet)",
    "players": 
    {
        "max": 10,
        "online": 0
    }
}
`

func handleServerListPing(clientConn net.Conn, app *model.App) {
	logging.Info("Player is pinging the server", "remoteAddr", clientConn.RemoteAddr().String())

	for {
		// receive packet
		packet, err := mcp.ReadPacket(clientConn)
		if err != nil {
			logging.Error("Error reading packet:", "error", err)
			clientConn.Close()
			return
		}

		if packet.PacketID == 0x00 {
			statusResponse := defaultResponse

			if app.IsMCServerReady() {
				logging.Info("MC Server is ready, sending status response")
				status, err := mc.ReadServerStatus("91.107.192.240", app.Config.MC_PORT)
				if err != nil {
					logging.Error("Error reading server status:", "error", err)
				} else {
					statusResponse = status
				}
			}

			mc.WriteStatusResponse(clientConn, statusResponse)
		} else if packet.PacketID == 0x01 {
			mc.WritePongResponse(clientConn, packet.Data)

			clientConn.Close()
			return
		}
	}
}

func findPlayerUUID(playerName string, app *model.App) (string, error) {
	// Check if the player is in the database
	player, err := queries.GetPlayerByAlias(app.Db, playerName)
	uuid := player.McId

	if err != nil {
		logging.Info("Player not found in database, requesting UUID from Mojang", "playerName", playerName)

		uuid, err = mojang.GetPlayerUUID(playerName)

		if err != nil {
			logging.Error("Error getting player UUID from Mojang:", "error", err)
			return "", err
		}
	}

	logging.Info("Found UUID:", "uuid", uuid)

	return uuid, nil
}

func handlePlayerLogin(clientConn net.Conn, app *model.App) {
	logging.Info("Player is connecting to the server", "remoteAddr", clientConn.RemoteAddr().String())

	//Wait for Login packet
	packet, err := mcp.ReadPacket(clientConn)
	if err != nil {
		logging.Error("Error reading packet:", "error", err)
		clientConn.Close()
		return
	}
	loginStart, err := mc.ParseLoginStart(packet)
	if err != nil {
		logging.Error("Error reading login start:", "error", err)
		clientConn.Close()
		return
	}
	logging.Info("Player is logging in", "username", loginStart.Username, "uuid", loginStart.UUID)

	uuid := *loginStart.UUID
	if true || loginStart.UUID == nil {
		id_found, _ := findPlayerUUID(loginStart.Username, app)

		logging.Info("Player UUID found:", "uuid", id_found)
	}

	// TODO: Return if the player is not in the whitelist -- no query yet.
	/* player, err := queries.QueryPlayerByUUID(app, uuid)
	if err != nil {
		logging.Error("Error querying player:", "error", err)
		clientConn.Close()
		return
	}

	if !player.IsAllowed {
		logging.Error("Player is not allowed to connect", "username", loginStart.Username, "uuid", loginStart.UUID)
		clientConn.Close()
		return
	} */

	// Send LoginSuccess packet
	loginSuccess := mc.WriteLoginSuccess(clientConn, uuid, loginStart.Username)
	if loginSuccess != nil {
		logging.Error("Error writing login success:", "error", loginSuccess)
		clientConn.Close()
		return
	}
	logging.Info("Sent LoginSuccess", "username", loginStart.Username)

	// Wait for LoginAcknowledged packet
	packet, err = mcp.ReadPacket(clientConn)
	if err != nil {
		logging.Error("Error reading packet:", "error", err)
		clientConn.Close()
		return
	}
	loginAcknowledged, err := mc.ParseLoginAcknowledged(packet)
	if err != nil {
		logging.Error("Error reading login acknowledged:", "error", loginAcknowledged)
		clientConn.Close()
		return
	}
	logging.Info("Player acknowledged login", "loggedIn", loginAcknowledged, "username", loginStart.Username)

	for {
		if app.IsMCServerReady() {
			break
		}

		logging.Info("Server not ready, stalling player...")
		_, err = clientConn.Write([]byte{0x00})
		if err != nil {
			logging.Error("Error writing stall packet:", "error", err)
			clientConn.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}

	// Send Transfer packet
	err = mc.WriteTransfer(clientConn, "91.107.192.240", 25565)
	if err != nil {
		logging.Error("Error writing transfer:", "error", err)
		clientConn.Close()
		return
	}
	logging.Info("Sent Transfer packet", "host", "91.107.192.240", "port", 25565)
}

func handleLegacyPing(clientConn net.Conn, app *model.App) {
	mc.WriteLegacyStatusResponse(clientConn, "127", "1.21.5", "Disco, Disco! Party, Party! (not started yet)", "0", "10")
}

// Handle incoming player connections
func handleConnection(clientConn net.Conn, app *model.App) {
	logging.Info("Player attempting to connect...", "remoteAddr", clientConn.RemoteAddr().String())
	packet, err := mcp.ReadPacket(clientConn)
	if err != nil {
		logging.Error("Error reading packet:", "error", err)
		clientConn.Close()
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
		logging.Error("Error reading handshake:", "error", err)
		clientConn.Close()
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
		return
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
