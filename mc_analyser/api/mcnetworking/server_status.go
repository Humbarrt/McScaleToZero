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

	WriteHandshake(conn, 770, host, port, 1)

	// Send Request packet
	WriteStatusRequest(conn)

	p, err := mcp.ReadPacket(conn)
	if err != nil {
		logging.Error("Error reading handshake:", "error", err)
		conn.Close()
		return "", err
	}

	status, err := ParseStatusResponse(p)
	conn.Close()

	return status, err
}
