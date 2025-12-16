package mc_networking

import (
	"fmt"
	"mc_manager/api/mc_networking/packet"
	"mc_manager/logging"
	"mc_manager/model"
)

type HandshakePacket struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	NextState       int
}

func ParseHandshake(p *packet.Packet) (HandshakePacket, error) {
	if p.PacketID != 0x00 {
		logging.Error("Invalid packet ID for HandshakePacket:", "expected", 0x00, "got", p.PacketID)
		return HandshakePacket{}, fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}
	var err error
	var h HandshakePacket
	// Read the HandshakePacket data
	h.ProtocolVersion, err = packet.ConsumeVarInt(&p.Data)
	if err != nil {
		logging.Error("Error reading protocol version:", "error", err)
		return HandshakePacket{}, err
	}

	h.ServerAddress, err = packet.ConsumeString(&p.Data)
	if err != nil {
		logging.Error("Error reading server address:", "error", err)
		return HandshakePacket{}, err
	}

	h.ServerPort, err = packet.ConsumeUnsignedShort(&p.Data)
	if err != nil {
		logging.Error("Error reading server port:", "error", err)
		return HandshakePacket{}, err
	}

	h.NextState, err = packet.ConsumeVarInt(&p.Data)
	if err != nil {
		logging.Error("Error reading next state:", "error", err)
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
		logging.Error("Invalid packet ID for login start:", "expected", 0x00, "got", p.PacketID)
		return LoginStartPacket{}, fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}

	var err error
	var ls LoginStartPacket
	ls.Username, err = packet.ConsumeString(&p.Data)
	if err != nil {
		logging.Error("Error reading username:", "error", err)
		return LoginStartPacket{}, err
	}
	if len(p.Data) == 0 {
		return ls, nil
	}

	uuid, err := packet.ConsumeUUID(&p.Data)
	if err != nil {
		logging.Error("Error reading UUID:", "error", err)
		return LoginStartPacket{}, err
	}
	ls.UUID = &uuid
	return ls, nil
}

func ParseLoginAcknowledged(p *packet.Packet) (bool, error) {
	if p.PacketID != 0x03 {
		return false, fmt.Errorf("invalid packet ID: expected 0x03, got %d", p.PacketID)
	}
	return true, nil
}

func ParseKeepAlive(p *packet.Packet) (int64, error) {
	if p.PacketID != 0x04 {
		return 0, fmt.Errorf("invalid packet ID: expected 0x04, got %d", p.PacketID)
	}
	return packet.ConsumeLong(&p.Data)
}

func ParseStatusResponse(p *packet.Packet) (string, error) {
	if p.PacketID != 0x00 {
		logging.Error("Invalid packet ID for StatusResponse:", "expected", 0x00, "got", p.PacketID)
		return "", fmt.Errorf("invalid packet ID: expected 0x00, got %d", p.PacketID)
	}
	return packet.ConsumeString(&p.Data)
}

type EncryptionRequestPacket struct {
	SharedSecret packet.PrefixedArray
	VerifyToken  packet.PrefixedArray
}

func ParseEncryptionResponse(p *packet.Packet) (EncryptionRequestPacket, error) {
	if p.PacketID != 0x01 {
		logging.Error("Invalid packet ID for EncryptionResponse:", "expected", 0x01, "got", p.PacketID)
		return EncryptionRequestPacket{}, fmt.Errorf("invalid packet ID: expected 0x01, got %d", p.PacketID)
	}

	var err error
	var erp EncryptionRequestPacket
	erp.SharedSecret, err = packet.ConsumePrefixedArray(&p.Data)
	if err != nil {
		logging.Error("Error Reading SharedSecret:", "error", err)
		return erp, err
	}

	erp.VerifyToken, err = packet.ConsumePrefixedArray(&p.Data)
	if err != nil {
		logging.Error("Error Reading VerifyToken:", "error", err)
		return erp, err
	}
	if len(p.Data) != 0 {
		logging.Error("Extra data in packet:", "data", p.Data)
		return erp, fmt.Errorf("extra data in packet")
	}

	logging.Debug("Parsed EncryptionRequestPacket", "SharedSecret", erp.SharedSecret, "VerifyToken", erp.VerifyToken)
	return erp, nil
}
