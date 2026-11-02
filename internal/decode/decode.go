package decode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

var opCodes = map[byte]string{
	0b100010: "mov",
}

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

func DecodeInstructions(data []byte) (string, error) {
	var decodeBuilder strings.Builder
	decodeBuilder.WriteString("bits 16\n")

	for i := 0; i < len(data); i += 1 {
		instruction := ""

		numBytes := 2

		byte1 := data[i]
		byte2 := data[i+1]

		opCode := byte1 >> 2
		d := byte1 & 0b00000010 >> 1
		w := byte1 & 0b00000001
		mod := byte2 >> 6
		reg := byte2 & 0b00111000 >> 3
		rm := byte2 & 0b00000111

		opCodeString, ok := opCodes[opCode]
		if !ok {
			return "", fmt.Errorf("Invalid opcode: %b", byte1>>2)
		}
		instruction += fmt.Sprintf("%s ", opCodeString)

		switch mod {
		case 0b00:
			if rm == 0b110 {
				numBytes += 2
				d16 := binary.LittleEndian.Uint16([]byte{data[i+2], data[i+3]})
				if d == 0 {
					instruction += fmt.Sprintf("[%d],%s", d16, registers[w][reg])
				} else {
					instruction += fmt.Sprintf("%s,[%d]", registers[w][reg], d16)
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
				d8 := int(data[i+2])
				if d == 0 {
					instruction += fmt.Sprintf("[%s+%d],%s", effectiveAddressCalculations[rm], d8, registers[w][reg])
				} else {
					instruction += fmt.Sprintf("%s,[%s+%d]", registers[w][reg], effectiveAddressCalculations[rm], d8)
				}
			}
		case 0b10:
			{
				numBytes += 2
				d16 := binary.LittleEndian.Uint16([]byte{data[i+2], data[i+3]})
				if d == 0 {
					instruction += fmt.Sprintf("[%s+%d],%s", effectiveAddressCalculations[rm], d16, registers[w][reg])
				} else {
					instruction += fmt.Sprintf("%s,[%s+%d]", registers[w][reg], effectiveAddressCalculations[rm], d16)
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

		fmt.Fprintf(&decodeBuilder, "\n%s", instruction)

		i += numBytes - 1
	}

	return decodeBuilder.String(), nil
}
