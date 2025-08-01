package packet

import (
	"fmt"
	"mc_manager/logging"
	"net"
	"time"
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

	//logging.Debug("Packet bytes", "packetID", p.PacketID, "length", p.Length, "data", packetBytes)
	return packetBytes
}

func readPacketLength(conn net.Conn) (int, error) {
	var num int
	var numRead int

	var byteRead byte
	maxPacketSize := 10 * 1024 // 10 KB
	startedAt := time.Now()
	for {
		if num > maxPacketSize {
			return 0, fmt.Errorf("packet size exceeds maximum limit")
		}

		if time.Since(startedAt) > 1*time.Second {
			return 0, fmt.Errorf("timeout while reading packet length")
		}

		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			return 0, err
		}
		byteRead = buf[0]
		if byteRead == 0x00 {
			logging.Debug("Got stalling byte")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		num |= int(byteRead&0x7F) << (7 * numRead)
		numRead++
		if (byteRead & 0x80) == 0 {
			break
		}
	}
	return num, nil
}

type PrefixedArray struct {
	Length int
	Data   []byte
}

type PrefixedOptional[T any] struct {
	IsPresent bool
	Value     T
}
