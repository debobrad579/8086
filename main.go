package main

import (
	"fmt"
	"os"

	"github.com/debobrad579/8086/internal/decode"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Invalid arguments")
		os.Exit(1)
	}

	filename := args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	decodedAsm, err := decode.DecodeInstructions(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(decodedAsm)
}
