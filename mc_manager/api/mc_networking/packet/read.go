package packet

import (
	"fmt"
	"mc_manager/logging"
	"mc_manager/model"
	"net"
)

func ReadPacket(conn net.Conn) (*Packet, error) {
	var p Packet
	// Read the packet length
	length, err := readPacketLength(conn)
	if err != nil {
		logging.Error("Error reading packet length:", "error", err)
		return &Packet{}, err
	}
	p.Length = length
	logging.Debug("Packet length", "length", length)

	// Read the packet data
	p.Data = make([]byte, length)
	totalBytesRead := 0
	// Read the rest of the packet
	for totalBytesRead < length {
		var bytesRead int
		bytesRead, err = conn.Read(p.Data[totalBytesRead:])
		totalBytesRead += bytesRead
	}

	if err != nil {
		logging.Error("Error reading packet data:", "error", err)
		return &Packet{}, err
	}

	// Read the packet ID
	p.PacketID, err = ConsumeVarInt(&p.Data)
	if err != nil {
		logging.Error("Error reading packet ID:", "error", err)
		return &Packet{}, err
	}

	return &p, nil
}

func ConsumeVarInt(data *[]byte) (int, error) {
	var num int
	var numRead int

	var byteRead byte
	for {
		if len(*data) == 0 {
			return 0, fmt.Errorf("not enough data to read")
		}
		byteRead = (*data)[0]
		*data = (*data)[1:]
		num |= int(byteRead&0x7F) << (7 * numRead)
		numRead++
		if (byteRead & 0x80) == 0 {
			break
		}
	}
	return num, nil
}

func ConsumeString(data *[]byte) (string, error) {
	length, err := ConsumeVarInt(data)
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if len(*data) < length {
		return "", fmt.Errorf("not enough data to read string")
	}
	copy(buf, (*data)[:length])
	*data = (*data)[length:]
	return string(buf), nil
}

func ConsumeUnsignedShort(data *[]byte) (uint16, error) {
	if len(*data) < 2 {
		return 0, fmt.Errorf("not enough data to read unsigned short")
	}
	value := uint16((*data)[0])<<8 | uint16((*data)[1])
	*data = (*data)[2:]
	return value, nil
}

func ConsumeUUID(data *[]byte) (model.UUID, error) {
	if len(*data) < 16 {
		return model.UUID{}, fmt.Errorf("not enough data to read UUID")
	}

	var uuid model.UUID
	copy(uuid.Bytes()[:], (*data)[:16])
	*data = (*data)[16:]
	return uuid, nil
}

func ConsumePrefixedArray(data *[]byte) (PrefixedArray, error) {
	if len(*data) == 0 {
		return PrefixedArray{}, fmt.Errorf("not enough data to read prefixed array")
	}
	var pa PrefixedArray
	length, err := ConsumeVarInt(data)
	if err != nil {
		return PrefixedArray{}, err
	}
	pa.Length = length
	if len(*data) < length {
		return PrefixedArray{}, fmt.Errorf("not enough data to read prefixed array")
	}

	pa.Data = (*data)[:length]
	*data = (*data)[length:]
	return pa, nil
}

func ConsumeItemFromPrefixedArray[T any](array *PrefixedArray, itemConsumer func(*[]byte) (T, error)) (T, error) {
	return itemConsumer(&array.Data)
}

func ConsumePrefixedOptional[T any](data *[]byte, valueConsumer func(*[]byte) (T, error)) (PrefixedOptional[T], error) {
	if len(*data) == 0 {
		return PrefixedOptional[T]{}, fmt.Errorf("not enough data to read prefixed optional")
	}
	var po PrefixedOptional[T]
	po.IsPresent = (*data)[0] == 1
	*data = (*data)[1:]

	if po.IsPresent {
		value, err := valueConsumer(data)
		if err != nil {
			return PrefixedOptional[T]{}, err
		}
		po.Value = value
	}
	return po, nil
}
