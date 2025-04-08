package mc_networking

import (
	"fmt"
	mcp "mc_analyser/api/mc_networking/packet"
	"mc_analyser/logging"
	"net"
	"time"
)

func ReadServerStatus(host string, port int) (string, error) {
	//connect to mc server via tcp
	conn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		logging.Error("Error opening tcp connection to mc-server:", "error", err)
		return "", err
	}
	defer conn.Close()

	// Send Handshake packet
	logging.Info("Sending handshake")
	err = WriteHandshake(conn, 770, host, port, 1)
	if err != nil {
		logging.Error("Error writing handshake:", "error", err)
		return "", err
	}

	// Send Request packet
	err = WriteStatusRequest(conn)
	if err != nil {
		logging.Error("Error writing status request:", "error", err)
		return "", err
	}

	// Read Response packet
	p, err := mcp.ReadPacket(conn)
	if err != nil {
		logging.Error("Error reading status response:", "error", err)
		return "", err
	}

	status, err := ParseStatusResponse(p)
	if err != nil {
		logging.Error("Error parsing status response:", "error", err)
		return "", err
	}

	return status, err
}

func IsServerTransferReady(host string, port int) (bool, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), 1*time.Second)
	if err != nil {
		logging.Error("Error opening tcp connection to mc-server:", "error", err)
		return false, err
	}
	defer conn.Close()
	logging.Debug("Server is ready for transfer")
	return true, nil
}
