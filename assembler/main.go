package main

import (
	"fmt"
	"os"
	"strings"
)

// [X, X, X, I, I, I, I, I]
// Leading 3 bits reserved for counter
const (
	NOP int8 = 0b00000000
	LDA int8 = 0b00000001
	ADD int8 = 0b00000010
	OUT int8 = 0b00011110
	HLT int8 = 0b00011111
)

func main() {
	instructions := make([]int8, 0)
	instructions = append(instructions, []int8{LDA, 10}...)
	instructions = append(instructions, []int8{ADD, 22}...)
	instructions = append(instructions, []int8{OUT, HLT, 0, 0, 0, 0, 20}...)
	writeRom(instructions)
}

func writeRom(prog []int8) {
	var sb strings.Builder
	sb.WriteString("v2.0 raw\n")
	f, err := os.OpenFile("prog.hex", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Error opening rom file %s\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()
	for i := 0; i < len(prog); i++ {
		sb.WriteString(fmt.Sprintf("%X", prog[i]))
		sb.WriteString("\n")
	}
	f.WriteString(sb.String())
}
