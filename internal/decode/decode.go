package decode

import (
	"encoding/binary"
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

func DecodeInstructions(encoded []byte) (string, error) {
	var decodeBuilder strings.Builder
	decodeBuilder.WriteString("bits 16\n")

	for i := 0; i < len(encoded); i += 1 {
		var instruction string

		numBytes := 2

		byte1 := encoded[i]
		byte2 := encoded[i+1]

		if byte1>>4 == 0b1011 {
			instruction = "mov "

			w := byte1 & 0b00001000 >> 3
			reg := byte1 & 0b00000111

			var byte3 byte
			if w == 1 {
				numBytes += 1
				byte3 = encoded[i+2]
			}

			instruction += fmt.Sprintf("%s,%d", registers[w][reg], bytesToData(byte2, byte3))
		} else if byte1>>2 == 0b100010 {
			instruction += "mov "

			d := byte1 & 0b00000010 >> 1
			w := byte1 & 0b00000001
			mod := byte2 >> 6
			reg := byte2 & 0b00111000 >> 3
			rm := byte2 & 0b00000111

			switch mod {
			case 0b00:
				if rm == 0b110 {
					numBytes += 2
					byte3 := encoded[i+2]
					byte4 := encoded[i+3]
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
					byte3 := encoded[i+2]
					if d == 0 {
						instruction += fmt.Sprintf("[%s%s],%s", effectiveAddressCalculations[rm], signed8BitDisplacement(byte3), registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%s%s]", registers[w][reg], effectiveAddressCalculations[rm], signed8BitDisplacement(byte3))
					}
				}
			case 0b10:
				{
					numBytes += 2
					byte3 := encoded[i+2]
					byte4 := encoded[i+3]
					if d == 0 {
						instruction += fmt.Sprintf("[%s%s],%s", effectiveAddressCalculations[rm], signedDisplacement(byte3, byte4), registers[w][reg])
					} else {
						instruction += fmt.Sprintf("%s,[%s%s]", registers[w][reg], effectiveAddressCalculations[rm], signedDisplacement(byte3, byte4))
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
			instruction = "mov "
			numBytes += 1

			w := byte1 & 0b00000001
			mod := byte2 >> 6
			rm := byte2 & 0b00000111

			size := "byte"
			if w == 1 {
				size = "word"
				numBytes += 1
			}

			switch mod {
			case 0b00:
				if rm == 0b110 {
					numBytes += 2
					byte3 := encoded[i+2]
					byte4 := encoded[i+3]
					byte5 := encoded[i+4]
					var byte6 byte
					if w == 1 {
						byte6 = encoded[i+5]
					}
					instruction += fmt.Sprintf("[%d],%s %d", bytesToData(byte3, byte4), size, bytesToData(byte5, byte6))
				} else {
					byte3 := encoded[i+2]
					var byte4 byte
					if w == 1 {
						byte4 = encoded[i+3]
					}
					instruction += fmt.Sprintf("[%s],%s %d", effectiveAddressCalculations[rm], size, bytesToData(byte3, byte4))
				}
			case 0b01:
				{
					numBytes += 1
					byte3 := encoded[i+2]
					byte4 := encoded[i+3]
					var byte5 byte
					if w == 1 {
						byte5 = encoded[i+4]
					}
					instruction += fmt.Sprintf("[%s%s],%s %d", effectiveAddressCalculations[rm], signed8BitDisplacement(byte3), size, bytesToData(byte4, byte5))
				}
			case 0b10:
				{
					numBytes += 2
					byte3 := encoded[i+2]
					byte4 := encoded[i+3]
					byte5 := encoded[i+4]
					var byte6 byte
					if w == 1 {
						byte6 = encoded[i+5]
					}
					instruction += fmt.Sprintf("[%s%s],%s %d", effectiveAddressCalculations[rm], signedDisplacement(byte3, byte4), size, bytesToData(byte5, byte6))
				}
			}
		} else if byte1>>2 == 0b101000 {
			instruction = "mov "
			numBytes += 1
			byte3 := encoded[i+2]

			d := byte1 & 0b00000010 >> 1
			_ = byte1 & 0b00000001

			addr := bytesToData(byte2, byte3)

			if d == 0 {
				instruction += fmt.Sprintf("AX,[%d]", addr)
			} else {
				instruction += fmt.Sprintf("[%d],AX", addr)
			}
		} else {
			return "", errors.New("Invalid instruction")
		}

		fmt.Fprintf(&decodeBuilder, "\n%s", instruction)

		i += numBytes - 1
	}

	return decodeBuilder.String(), nil
}

func bytesToData(byte1 byte, byte2 byte) uint16 {
	return binary.LittleEndian.Uint16([]byte{byte1, byte2})
}

func signedDisplacement(byte1 byte, byte2 byte) string {
	displacement := int16(bytesToData(byte1, byte2))
	var sign string
	if displacement >= 0 {
		sign = "+"
	}

	return fmt.Sprintf("%s%d", sign, displacement)
}

func signed8BitDisplacement(byte byte) string {
	displacement := int8(byte)
	var sign string
	if displacement >= 0 {
		sign = "+"
	}

	return fmt.Sprintf("%s%d", sign, displacement)
}
