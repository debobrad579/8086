package decode

import (
	"errors"
	"fmt"
	"strings"
)

var registers = [2]map[byte]string{
	{
		0b000: "AL",
		0b001: "CL",
		0b010: "DL",
		0b011: "BL",
		0b100: "AH",
		0b101: "CH",
		0b110: "DH",
		0b111: "BH",
	},
	{
		0b000: "AX",
		0b001: "CX",
		0b010: "DX",
		0b011: "BX",
		0b100: "SP",
		0b101: "BP",
		0b110: "SI",
		0b111: "DI",
	},
}

var effectiveAddressCalculations = map[byte]string{
	0b000: "BX+SI",
	0b001: "BX+DI",
	0b010: "BP+SI",
	0b011: "BP+DI",
	0b100: "SI",
	0b101: "DI",
	0b110: "BP",
	0b111: "BX",
}

var aluOpCodes = map[byte]string{
	0b000: "ADD",
	0b001: "OR",
	0b010: "ADC",
	0b011: "SBB",
	0b100: "AND",
	0b101: "SUB",
	0b110: "BP",
	0b111: "CMP",
}

func DecodeInstructions(encoded []byte) (string, error) {
	var decodeBuilder strings.Builder
	decodeBuilder.WriteString("bits 16\n")

	for i := 0; i < len(encoded); i += 1 {
		var instruction string

		numBytes := 2

		byte1 := getByteNumber(encoded, i, 1)
		byte2 := getByteNumber(encoded, i, 2)
		byte3 := getByteNumber(encoded, i, 3)
		byte4 := getByteNumber(encoded, i, 4)
		byte5 := getByteNumber(encoded, i, 5)
		byte6 := getByteNumber(encoded, i, 6)

		if byte1>>6 == 0b00 && (byte1&0b00000110)>>1 == 0b10 {
			instruction, numBytes = immediateToAccumulator(byte1, byte2, byte3)
		} else if byte1>>4 == 0b1011 {
			instruction = "MOV "

			w := byte1 & 0b00001000 >> 3
			reg := byte1 & 0b00000111

			if w == 1 {
				numBytes += 1
			}

			instruction += fmt.Sprintf("%s,%d", registers[w][reg], bytesToData(byte2, byte3))
		} else if byte1>>2 == 0b100010 || (byte1&0b11000100 == 0) {
			instruction, numBytes = regMemWithReg(byte1, byte2, byte3, byte4)
		} else if byte1>>1 == 0b1100011 || byte1>>2 == 0b100000 {
			instruction, numBytes = immediateToRegMem(byte1, byte2, byte3, byte4, byte5, byte6)
		} else if byte1>>2 == 0b101000 {
			instruction, numBytes = movAccumulator(byte1, byte2, byte3)
		} else {
			return "", errors.New("Invalid instruction")
		}

		fmt.Fprintf(&decodeBuilder, "\n%s", instruction)

		i += numBytes - 1
	}

	return decodeBuilder.String(), nil
}

func regMemWithReg(byte1 byte, byte2 byte, byte3 byte, byte4 byte) (instruction string, numBytes int) {
	if byte1>>2 == 0b100010 {
		instruction = "MOV"
	} else if byte1&0b11000100 == 0 {
		instruction = aluOpCodes[(byte1&0b00111000)>>3]
	}

	d := (byte1 & 0b00000010) >> 1
	w := byte1 & 0b00000001
	mod := byte2 >> 6
	reg := (byte2 & 0b00111000) >> 3
	rm := byte2 & 0b00000111

	switch mod {
	case 0b00:
		if rm == 0b110 {
			if d == 0 {
				return fmt.Sprintf(
					"%s [%d],%s",
					instruction,
					bytesToData(byte3, byte4),
					registers[w][reg],
				), 4
			} else {
				return fmt.Sprintf(
					"%s %s,[%d]",
					instruction,
					registers[w][reg],
					bytesToData(byte3, byte4),
				), 4
			}
		} else {
			if d == 0 {
				return fmt.Sprintf(
					"%s [%s],%s",
					instruction,
					effectiveAddressCalculations[rm],
					registers[w][reg],
				), 2
			} else {
				return fmt.Sprintf(
					"%s %s,[%s]",
					instruction,
					registers[w][reg],
					effectiveAddressCalculations[rm],
				), 2
			}
		}
	case 0b01:
		if d == 0 {
			return fmt.Sprintf(
				"%s [%s%+d],%s",
				instruction,
				effectiveAddressCalculations[rm],
				int8(byte3),
				registers[w][reg],
			), 3
		} else {
			return fmt.Sprintf(
				"%s %s,[%s%+d]",
				instruction,
				registers[w][reg],
				effectiveAddressCalculations[rm],
				int8(byte3),
			), 3
		}
	case 0b10:
		if d == 0 {
			return fmt.Sprintf(
				"%s [%s%+d],%s",
				instruction,
				effectiveAddressCalculations[rm],
				int16(bytesToData(byte3, byte4)),
				registers[w][reg],
			), 4
		} else {
			return fmt.Sprintf(
				"%s %s,[%s%+d]",
				instruction,
				registers[w][reg],
				effectiveAddressCalculations[rm],
				int16(bytesToData(byte3, byte4)),
			), 4
		}
	default:
		if d == 0 {
			return fmt.Sprintf(
				"%s %s,%s",
				instruction,
				registers[w][rm],
				registers[w][reg],
			), 2
		} else {
			return fmt.Sprintf(
				"%s %s,%s",
				instruction,
				registers[w][reg],
				registers[w][rm],
			), 2
		}
	}
}

func immediateToRegMem(
	byte1 byte,
	byte2 byte,
	byte3 byte,
	byte4 byte,
	byte5 byte,
	byte6 byte,
) (instruction string, numBytes int) {
	w := byte1 & 0b00000001
	s := (byte1 & 0b00000010) >> 1
	mod := byte2 >> 6
	opCode := (byte2 & 0b00111000) >> 3
	rm := byte2 & 0b00000111

	if byte1>>1 == 0b1100011 {
		instruction = "MOV"
		s = 0
	} else if byte1>>2 == 0b100000 {
		instruction = aluOpCodes[opCode]
	} else {
		return "", 0
	}

	var dataByte1 byte
	var dataByte2 byte
	var displacement string

	directAddressing := false

	switch mod {
	case 0b00:
		if rm == 0b110 {
			directAddressing = true
			displacement = fmt.Sprintf(
				"%d",
				bytesToData(byte3, byte4),
			)
			dataByte1 = byte5
			dataByte2 = byte6
			numBytes = 4
		} else {
			dataByte1 = byte3
			dataByte2 = byte4
			numBytes = 2
		}
	case 0b01:
		displacement = fmt.Sprintf("%+d", int8(byte3))
		dataByte1 = byte4
		dataByte2 = byte5
		numBytes = 3
	case 0b10:
		displacement = fmt.Sprintf(
			"%+d",
			int16(bytesToData(byte3, byte4)),
		)
		dataByte1 = byte5
		dataByte2 = byte6
		numBytes = 4
	case 0b11:
		dataByte1 = byte3
		dataByte2 = byte4
		numBytes = 2
	}

	immediateSize := 1

	if w == 1 && s == 0 {
		immediateSize = 2
	}

	var data string
	var size string

	if w == 0 {
		size = "byte"
		data = fmt.Sprintf("%d", dataByte1)
	} else {
		size = "word"

		if s == 1 {
			data = fmt.Sprintf("%d", int8(dataByte1))
		} else {
			data = fmt.Sprintf(
				"%d",
				int16(bytesToData(dataByte1, dataByte2)),
			)
		}
	}

	numBytes += immediateSize

	if mod == 0b11 {
		return fmt.Sprintf(
			"%s %s,%s %s",
			instruction,
			registers[w][rm],
			size,
			data,
		), numBytes
	}

	if directAddressing {
		return fmt.Sprintf(
			"%s [%s],%s %s",
			instruction,
			displacement,
			size,
			data,
		), numBytes
	}

	return fmt.Sprintf(
		"%s [%s%s],%s %s",
		instruction,
		effectiveAddressCalculations[rm],
		displacement,
		size,
		data,
	), numBytes
}

func movAccumulator(byte1 byte, byte2 byte, byte3 byte) (instruction string, numBytes int) {
	d := (byte1 & 0b00000010) >> 1
	w := byte1 & 0b00000001

	register := "AX"
	if w == 0 {
		register = "AL"
	}

	addr := bytesToData(byte2, byte3)

	if d == 0 {
		return fmt.Sprintf("MOV %s,[%d]", register, addr), 3
	} else {
		return fmt.Sprintf("MOV [%d],%s", addr, register), 3
	}
}

func immediateToAccumulator(byte1 byte, byte2 byte, byte3 byte) (instruction string, numBytes int) {
	instruction = aluOpCodes[(byte1&0b00111000)>>3]

	w := byte1 & 0b00000001

	if w == 0 {
		return fmt.Sprintf("%s AL,%d", instruction, int8(byte2)), 2
	} else {
		return fmt.Sprintf("%s AX,%d", instruction, int16(bytesToData(byte2, byte3))), 3
	}
}
