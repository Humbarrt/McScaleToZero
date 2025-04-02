package mcnetworking

import (
	"fmt"
	mcp "mc_analyser/api/mcnetworking/packet"
	"mc_analyser/logging"
	"net"
)

func ReadServerStatus(host string, port int) (string, error) {
	//connect to mc server via tcp
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return "", err
	}

	err = WriteHandshake(conn, 770, host, port, 1)
	if err != nil {
		logging.Error("Error writing handshake:", "error", err)
		conn.Close()
		return "", err
	}

	// Send Request packet
	err = WriteStatusRequest(conn)
	if err != nil {
		logging.Error("Error writing status request:", "error", err)
		conn.Close()
		return "", err
	}

	p, err := mcp.ReadPacket(conn)
	if err != nil {
		logging.Error("Error reading status response:", "error", err)
		conn.Close()
		return "", err
	}

	status, err := ParseStatusResponse(p)
	if err != nil {
		logging.Error("Error parsing status response:", "error", err)
		conn.Close()
		return "", err
	}
	conn.Close()

	return status, err
}
