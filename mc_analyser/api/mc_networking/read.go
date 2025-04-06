package mc_networking

import (
	"fmt"
	"log"
	"mc_analyser/api/mc_networking/packet"
	"mc_analyser/model"
)

type HandshakePacket struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	NextState       int
}

func ParseHandshake(p *packet.Packet) (HandshakePacket, error) {
	if p.PacketID != 0x00 {
		log.Println("Invalid packet ID for HandshakePacket:", "expected", 0x00, "got", p.PacketID)
		return HandshakePacket{}, fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}
	var err error
	var h HandshakePacket
	// Read the HandshakePacket data
	h.ProtocolVersion, err = packet.ConsumeVarInt(&p.Data)
	if err != nil {
		log.Println("Error reading protocol version:", "error", err)
		return HandshakePacket{}, err
	}

	h.ServerAddress, err = packet.ConsumeString(&p.Data)
	if err != nil {
		log.Println("Error reading server address:", "error", err)
		return HandshakePacket{}, err
	}

	h.ServerPort, err = packet.ConsumeUnsignedShort(&p.Data)
	if err != nil {
		log.Println("Error reading server port:", "error", err)
		return HandshakePacket{}, err
	}

	h.NextState, err = packet.ConsumeVarInt(&p.Data)
	if err != nil {
		log.Println("Error reading next state:", "error", err)
		return HandshakePacket{}, err
	}
	return h, nil
}

type LoginStartPacket struct {
	Username string
	UUID     *model.UUID
}

func ParseLoginStart(p *packet.Packet) (LoginStartPacket, error) {
	if p.PacketID != 0x00 {
		log.Println("Invalid packet ID for login start:", "expected", 0x00, "got", p.PacketID)
		return LoginStartPacket{}, fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}

	var err error
	var ls LoginStartPacket
	ls.Username, err = packet.ConsumeString(&p.Data)
	if err != nil {
		log.Println("Error reading username:", "error", err)
		return LoginStartPacket{}, err
	}
	if len(p.Data) == 0 {
		return ls, nil
	}

	uuid, err := packet.ConsumeUUID(&p.Data)
	if err != nil {
		log.Println("Error reading UUID:", "error", err)
		return LoginStartPacket{}, err
	}
	ls.UUID = &uuid

	log.Println("LoginStart UUID:", "uuid", ls.UUID.LongString())
	return ls, nil
}

func ParseLoginAcknowledged(p *packet.Packet) (bool, error) {
	if p.PacketID != 0x03 {
		return false, fmt.Errorf("invalid packet ID: expected 0x03, got %d", p.PacketID)
	}
	return true, nil
}

func ParseStatusResponse(p *packet.Packet) (string, error) {
	if p.PacketID != 0x00 {
		log.Println("Invalid packet ID for StatusResponse:", "expected", 0x00, "got", p.PacketID)
		return "", fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}
	return packet.ConsumeString(&p.Data)
}
