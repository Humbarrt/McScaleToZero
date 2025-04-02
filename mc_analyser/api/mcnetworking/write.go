package mcnetworking

import (
	"mc_analyser/api/mcnetworking/packet"
	"net"
	"unicode/utf16"
)

func WriteLoginSuccess(conn net.Conn, uuid packet.UUID, username string) error {
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
	data := make([]byte, 0)
	packet.WriteString(&data, host)
	packet.WriteVarInt(&data, port)
	p.Data = data

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
	data := make([]byte, 0)
	packet.WriteString(&data, status)
	p.Data = data

	packetBytes := p.ToBytes()
	conn.Write(packetBytes)
}

func WriteHandshake(conn net.Conn, protocolVersion int, serverAddress string, serverPort int, nextState int) error {
	p := packet.Packet{
		PacketID: 0x00,
		Data:     make([]byte, 0),
	}
	data := make([]byte, 0)
	packet.WriteVarInt(&data, protocolVersion)
	packet.WriteString(&data, serverAddress)
	packet.WriteUnsignedShort(&data, uint16(serverPort))
	packet.WriteVarInt(&data, nextState)
	p.Data = data

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

func stringToBytesUTF16BE(str string) []byte {
	codes := utf16.Encode([]rune(str))
	b := make([]byte, len(codes)*2)
	for i, r := range codes {
		b[i*2] = byte(r >> 8)
		b[i*2+1] = byte(r)
	}
	return b
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
