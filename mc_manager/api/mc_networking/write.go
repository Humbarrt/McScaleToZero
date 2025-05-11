package mc_networking

import (
	"encoding/json"
	"mc_manager/api/mc_networking/packet"
	"mc_manager/logging"
	"mc_manager/model"
	"net"
	"unicode/utf16"
)

func WriteLoginSuccess(conn net.Conn, uuid model.UUID, username string) error {
	p := packet.Packet{
		PacketID: 0x02,
		Data:     make([]byte, 0),
	}
	data := make([]byte, 0)
	packet.WriteUUID(&data, uuid)
	packet.WriteString(&data, username)

	var property packet.PrefixedArray
	packet.WritePrefixedArray(&data, &property)
	p.Data = data

	packetBytes := p.ToBytes()
	_, err := conn.Write(packetBytes)
	return err
}

func WriteTransfer(conn net.Conn, host string, port int) error {
	p := packet.Packet{
		PacketID: 0x0B,
		Data:     make([]byte, 0),
	}
	packet.WriteString(&p.Data, host)
	packet.WriteVarInt(&p.Data, port)

	packetBytes := p.ToBytes()
	_, err := conn.Write(packetBytes)
	return err
}

func WritePongResponse(conn net.Conn, Data []byte) error {
	p := packet.Packet{
		PacketID: 0x01,
		Data:     Data,
	}

	packetBytes := p.ToBytes()
	_, err := conn.Write(packetBytes)
	return err
}

func WriteStatusResponse(conn net.Conn, status string) {
	p := packet.Packet{
		PacketID: 0x00,
		Data:     make([]byte, 0),
	}

	packet.WriteString(&p.Data, status)

	packetBytes := p.ToBytes()
	conn.Write(packetBytes)
}

func WriteHandshake(conn net.Conn, protocolVersion int, serverAddress string, serverPort int, nextState int) error {
	p := packet.Packet{
		PacketID: 0x00,
		Data:     make([]byte, 0),
	}
	packet.WriteVarInt(&p.Data, protocolVersion)
	packet.WriteString(&p.Data, serverAddress)
	packet.WriteUnsignedShort(&p.Data, uint16(serverPort))
	packet.WriteVarInt(&p.Data, nextState)
	logging.Info("Handshake packet data: ", "data", p.Data)

	packetBytes := p.ToBytes()
	_, err := conn.Write(packetBytes)
	return err
}

func WriteStatusRequest(conn net.Conn) error {
	p := packet.Packet{
		PacketID: 0x00,
		Data:     make([]byte, 0),
	}

	packetBytes := p.ToBytes()
	_, err := conn.Write(packetBytes)
	return err
}

func WriteEncryptionRequest(conn model.McConn) error {
	p := packet.Packet{
		PacketID: 0x01,
		Data:     make([]byte, 0),
	}

	serverID := ""
	packet.WriteString(&p.Data, serverID)

	encryption, err := conn.GetEncryption()
	if err != nil {
		logging.Error("Error getting encryption:", "error", err)
		return err
	}

	var publicKeyArray packet.PrefixedArray
	publicKeyArray.Data = encryption.GetPublicKey()
	packet.WritePrefixedArray(&p.Data, &publicKeyArray)

	var verifyTokenArray packet.PrefixedArray
	verifyTokenArray.Data = encryption.GetVerifyToken()
	packet.WritePrefixedArray(&p.Data, &verifyTokenArray)

	shouldAuthenticate := true
	packet.WriteBoolean(&p.Data, shouldAuthenticate)

	packetBytes := p.ToBytes()
	_, err = conn.Write(packetBytes)
	return err
}

func WriteDisconnect(conn net.Conn, reason string) error {
	p := packet.Packet{
		PacketID: 0x00,
		Data:     make([]byte, 0),
	}

	jsonReason := map[string]string{
		"text":  reason,
		"color": "red",
	}
	jsonData, err := json.Marshal(jsonReason)
	if err != nil {
		logging.Error("Error marshalling disconnect data:", "error", err)
		return err
	}
	packet.WriteString(&p.Data, string(jsonData))

	packetBytes := p.ToBytes()
	_, err = conn.Write(packetBytes)
	return err
}

func WriteLegacyStatusResponse(conn net.Conn, protocolVersion string, minecraftVersion string, motd string, playerCount string, maxPlayers string) error {
	data := make([]byte, 0)

	data = append(data, stringToBytesUTF16BE("§1")...)
	data = append(data, 0x00, 0x00)

	data = append(data, stringToBytesUTF16BE(protocolVersion)...)
	data = append(data, 0x00, 0x00)

	data = append(data, stringToBytesUTF16BE(minecraftVersion)...)
	data = append(data, 0x00, 0x00)

	data = append(data, stringToBytesUTF16BE(motd)...)
	data = append(data, 0x00, 0x00)

	data = append(data, stringToBytesUTF16BE(playerCount)...)
	data = append(data, 0x00, 0x00)

	data = append(data, stringToBytesUTF16BE(maxPlayers)...)

	length := len(data) / 2

	packetBytes := make([]byte, 0)
	packetBytes = append(packetBytes, 0xFF)
	packetBytes = append(packetBytes, byte(length>>8), byte(length))
	packetBytes = append(packetBytes, data...)

	_, err := conn.Write(packetBytes)
	return err
}

func stringToBytesUTF16BE(str string) []byte {
	codes := utf16.Encode([]rune(str))
	b := make([]byte, len(codes)*2)
	for i, r := range codes {
		b[i*2] = byte(r >> 8)
		b[i*2+1] = byte(r)
	}
	return b
}
