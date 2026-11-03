package decode

import "encoding/binary"

func bytesToData(byte1 byte, byte2 byte) uint16 {
	return binary.LittleEndian.Uint16([]byte{byte1, byte2})
}

func getByteNumber(encoded []byte, i int, num int) byte {
	index := i + num - 1
	if index < 0 || index >= len(encoded) {
		return 0
	}
	return encoded[index]
}
