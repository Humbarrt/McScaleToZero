package networking

import (
	"mc_proxy/mc_net/packet"
	"net"
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
