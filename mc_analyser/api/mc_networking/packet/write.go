package packet

import "mc_analyser/model"

func WriteVarInt(data *[]byte, num int) {
	for {
		if (num & ^0x7F) == 0 {
			*data = append(*data, byte(num))
			break
		} else {
			*data = append(*data, byte(num&0x7F|0x80))
			num >>= 7
		}
	}
}

func WriteUnsignedShort(data *[]byte, num uint16) {
	*data = append(*data, byte(num>>8), byte(num))
}

func WriteString(data *[]byte, str string) {
	WriteVarInt(data, len(str))
	*data = append(*data, str...)
}

func WriteUUID(data *[]byte, uuid model.UUID) {
	*data = append(*data, uuid.Bytes()[:]...)
}

func WritePrefixedArray(data *[]byte, array *PrefixedArray) {
	WriteVarInt(data, len(array.Data))
	*data = append(*data, array.Data...)
}

func WritePrefixedArrayItem[T any](array *PrefixedArray, item []byte) {
	array.Data = append(array.Data, item...)
}

func WritePrefixedOptional[T any](data *[]byte, po PrefixedOptional[T], writeFunc func(data *[]byte, value T)) {
	if po.IsPresent {
		WriteVarInt(data, 1)
		writeFunc(data, po.Value)
	} else {
		WriteVarInt(data, 0)
	}
}
