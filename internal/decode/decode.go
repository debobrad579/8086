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
		} else if byte1>>2 == 0b100010 {
			instruction += "MOV "

			d := byte1 & 0b00000010 >> 1
			w := byte1 & 0b00000001
			mod := byte2 >> 6
			reg := byte2 & 0b00111000 >> 3
			rm := byte2 & 0b00000111

			switch mod {
			case 0b00:
				if rm == 0b110 {
					numBytes += 2
					if d == 0 {
						instruction += fmt.Sprintf("[%d],%s", bytesToData(byte3, byte4), registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%d]", registers[w][reg], bytesToData(byte3, byte4))
					}
				} else {
					if d == 0 {
						instruction += fmt.Sprintf("[%s],%s", effectiveAddressCalculations[rm], registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%s]", registers[w][reg], effectiveAddressCalculations[rm])
					}
				}
			case 0b01:
				{
					numBytes += 1
					if d == 0 {
						instruction += fmt.Sprintf("[%s%+d],%s", effectiveAddressCalculations[rm], int8(byte3), registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%s%+d]", registers[w][reg], effectiveAddressCalculations[rm], int8(byte3))
					}
				}
			case 0b10:
				{
					numBytes += 2
					if d == 0 {
						instruction += fmt.Sprintf("[%s%+d],%s", effectiveAddressCalculations[rm], int16(bytesToData(byte3, byte4)), registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%s%+d]", registers[w][reg], effectiveAddressCalculations[rm], int16(bytesToData(byte3, byte4)))
					}
				}
			case 0b11:
				{
					if d == 0 {
						instruction += fmt.Sprintf("%s,%s", registers[w][rm], registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,%s", registers[w][reg], registers[w][rm])
					}
				}
			}
		} else if byte1>>1 == 0b1100011 {
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

func immediateToRegMem(
	byte1 byte,
	byte2 byte,
	byte3 byte,
	byte4 byte,
	byte5 byte,
	byte6 byte,
) (instruction string, numBytes int) {
	w := byte1 & 0b00000001
	mod := byte2 >> 6
	rm := byte2 & 0b00000111

	if byte1>>1 == 0b1100011 {
		instruction = "MOV"
	}

	size := "byte"
	if w == 1 {
		size = "word"
	}

	switch mod {
	case 0b00:
		if rm == 0b110 {
			return fmt.Sprintf(
				"%s [%d],%s %d",
				instruction,
				bytesToData(byte3, byte4),
				size,
				bytesToData(byte5, byte6),
			), 5 + int(w)
		} else {
			return fmt.Sprintf(
				"%s [%s],%s %d",
				instruction,
				effectiveAddressCalculations[rm],
				size,
				bytesToData(byte3, byte4),
			), 3 + int(w)
		}
	case 0b01:
		return fmt.Sprintf(
			"%s [%s%+d],%s %d",
			instruction,
			effectiveAddressCalculations[rm],
			int8(byte3),
			size,
			bytesToData(byte4, byte5),
		), 4 + int(w)
	case 0b10:
		return fmt.Sprintf("%s [%s%+d],%s %d",
			instruction,
			effectiveAddressCalculations[rm],
			int16(bytesToData(byte3, byte4)),
			size,
			bytesToData(byte5, byte6),
		), 5 + int(w)
	default:
		return "", 0
	}
}

func movAccumulator(byte1 byte, byte2 byte, byte3 byte) (instruction string, numBytes int) {
	d := byte1 & 0b00000010 >> 1
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
		return fmt.Sprintf("%s AL,%d", instruction, int(byte2)), 2
	} else {
		return fmt.Sprintf("%s AX,%d", instruction, int(bytesToData(byte2, byte3))), 3
	}
}
