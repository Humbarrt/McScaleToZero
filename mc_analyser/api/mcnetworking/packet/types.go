package packet

import (
	"fmt"
	"net"
)

type Packet struct {
	Length   int
	PacketID int
	Data     []byte
}

func (p *Packet) CalcLength() {
	p.Length = len(p.Data) + 1 // +1 for the packet ID
}

func (p *Packet) ToBytes() []byte {
	p.CalcLength()
	packetBytes := make([]byte, 0)
	WriteVarInt(&packetBytes, p.Length)
	WriteVarInt(&packetBytes, p.PacketID)
	packetBytes = append(packetBytes, p.Data...)
	return packetBytes
}

func readPacketLength(conn net.Conn) (int, error) {
	var num int
	var numRead int

	var byteRead byte
	for {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			return 0, err
		}
		byteRead = buf[0]
		num |= int(byteRead&0x7F) << (7 * numRead)
		numRead++
		if (byteRead & 0x80) == 0 {
			break
		}
	}
	return num, nil
}

type UUID struct {
	uuid [16]byte
}

func (u *UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u.uuid[0:4],
		u.uuid[4:6],
		u.uuid[6:8],
		u.uuid[8:10],
		u.uuid[10:])
}

type PrefixedArray struct {
	Length int
	Data   []byte
}

type PrefixedOptional[T any] struct {
	IsPresent bool
	Value     T
}
