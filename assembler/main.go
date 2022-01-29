package main

import (
	"fmt"
	"os"
	"strings"
)

// [X, X, X, I, I, I, I, I]
// Leading 3 bits reserved for counter
const (
	NOP  int8 = 0b00000000 // No argument, do nothing
	LDA  int8 = 0b00000001 // Single argument, load the value from memory address $0 into A
	LDAi int8 = 0b00000010 // Single argument, store $0 in A
	LDB  int8 = 0b00000011 // Single argument, load the value from memory address $0 into A
	LDBi int8 = 0b00000100 // Single argument, store $0 in B
	STA  int8 = 0b00000101 // Single argument, store the value from A into memory address $0
	ADD  int8 = 0b00000110 // No argument, add A and B, store the result in A
	SUB  int8 = 0b00000111 // No argument, subtract A and B, store the result in A
	JMP  int8 = 0b00001000 // Single argument, unconditional jump to address $0
	JZ   int8 = 0b00001001 // Single argument, jump to address $0 if register A = 0
	JEQ  int8 = 0b00001010 // Single argument, jump to address $0 if register A = B
	JGE  int8 = 0b00001011 // Single argument, jump to address $0 if register A >= B
	OUT  int8 = 0b00011110 // No argument, display the value stored in register A
	HLT  int8 = 0b00011111 // No argument, halt the CPU
)

func main() {
	instructions := make([]int8, 0)
	/*instructions = append(instructions, []int8{LDA, 10}...)
	instructions = append(instructions, []int8{ADD, 22}...)
	instructions = append(instructions, []int8{OUT, HLT, 0, 0, 0, 0, 20}...)
	*/
	instructions = append(instructions, []int8{LDAi, 5}...)
	instructions = append(instructions, []int8{OUT}...)
	instructions = append(instructions, []int8{LDBi, 1}...)
	instructions = append(instructions, []int8{SUB}...)
	instructions = append(instructions, []int8{OUT}...)
	instructions = append(instructions, []int8{JZ, 11}...)
	instructions = append(instructions, []int8{JMP, 3}...)
	instructions = append(instructions, []int8{HLT}...)
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
