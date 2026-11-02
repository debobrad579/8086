package decode

import (
	"errors"
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

func DecodeInstructions(data []byte) (string, error) {
	var decodeBuilder strings.Builder
	decodeBuilder.WriteString("bits 16\n")

	for i := 0; i < len(data); i += 2 {
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

		if mod != 0b11 {
			return "", errors.New("Invalid mode")
		}

		regString := registers[w][reg]
		rmString := registers[w][rm]

		if d == 0 {
			fmt.Fprintf(&decodeBuilder, "\n%s %s,%s", opCodeString, rmString, regString)
		} else {
			fmt.Fprintf(&decodeBuilder, "\n%s %s,%s", opCodeString, regString, rmString)
		}
	}

	return decodeBuilder.String(), nil
}
