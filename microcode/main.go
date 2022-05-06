package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type ControlLine int64
type Instruction int64

const (
	PCOut ControlLine = iota
	PCEnable
	PCClear
	PCLd
	ALUOut
	LdOut
	LdA
	EnA
	LdB
	EnB
	Sub
	LdInst
	EnInst
	LdMAddr
	MWr
	MEn
	Hlt
	LdFlags
	EnGnd
	RstDone
	MPCRst
	PCLdIfEq
	PCLdIfGorEq
	PCLdIfZero
	StkEn
	StkLd
	StkInc
	StkDec
	StkRst
)

const (
	// Addressable ROM
	addressBits     = 8
	pcBits          = 3
	instructionBits = 5
	// Determines control width, update if new control lines are added
	highestLine = StkRst
)

// [X, X, X, I, I, I, I, I]
// Leading 3 bits reserved for counter
const (
	NOP  Instruction = 0b00000000 // No argument, do nothing
	LDA  Instruction = 0b00000001 // Single argument, load the value from memory address $0 into A
	LDAi Instruction = 0b00000010 // Single argument, store $0 in A
	LDB  Instruction = 0b00000011 // Single argument, load the value from memory address $0 into A
	LDBi Instruction = 0b00000100 // Single argument, store $0 in B
	STA  Instruction = 0b00000101 // Single argument, store the value from A into memory address $0
	ADD  Instruction = 0b00000110 // No argument, add A and B, store the result in A
	SUB  Instruction = 0b00000111 // No argument, subtract A and B, store the result in A
	JMP  Instruction = 0b00001000 // Single argument, unconditional jump to address $0
	JZ   Instruction = 0b00001001 // Single argument, jump to address $0 if register A = 0
	JEQ  Instruction = 0b00001010 // Single argument, jump to address $0 if register A = B
	JGE  Instruction = 0b00001011 // Single argument, jump to address $0 if register A >= B
	PUSH Instruction = 0b00001100 // No argument, push the A register onto the stack
	POP  Instruction = 0b00001101 // No argument, pop the top of the stack into the A register
	CALL Instruction = 0b00001110 // Single argument, push the program counter onto the stack and jump to address $0
	RET  Instruction = 0b00001111 // No argument, pop the top of the stack and jump to that address
	MOVa Instruction = 0b00010000 // No arguments, move B into A
	MOVb Instruction = 0b00010001 // No arguments, move A into B
	RST  Instruction = 0b00011101 // Reset the CPU
	OUT  Instruction = 0b00011110 // No argument, display the value stored in register A
	HLT  Instruction = 0b00011111 // No argument, halt the CPU
)

// These machine instructions get executed for every command
var preamble = [][]ControlLine{
	{PCOut, LdMAddr},
	{MEn, LdInst, PCEnable},
}

func main() {
	rom := make([][]ControlLine, int(math.Pow(2, float64(addressBits))))
	instructions := map[Instruction][][]ControlLine{
		NOP: {
			{PCEnable},
			{MPCRst},
		},
		LDA: {
			{PCOut, LdMAddr},
			{MEn, LdMAddr},
			{MEn, LdA, PCEnable},
			{MPCRst},
		},
		LDAi: {
			{PCOut, LdMAddr},
			{MEn, LdA, PCEnable},
			{MPCRst},
		},
		LDB: {
			{PCOut, LdMAddr},
			{MEn, LdMAddr},
			{MEn, LdB, PCEnable},
			{MPCRst},
		},
		LDBi: {
			{PCOut, LdMAddr},
			{MEn, LdB, PCEnable},
			{MPCRst},
		},
		STA: {
			{PCOut, LdMAddr},
			{MEn, LdMAddr},
			{EnA, MWr, PCEnable},
			{MPCRst},
		},
		ADD: {
			{ALUOut, LdA},
			{MPCRst},
		},
		SUB: {
			{ALUOut, Sub, LdA},
			{MPCRst},
		},
		JMP: {
			{PCOut, LdMAddr},
			{PCLd, MEn, PCEnable},
			{MPCRst},
		},
		JZ: {
			{LdFlags, PCOut, LdMAddr},
			{PCLd, PCLdIfZero, MEn, PCEnable},
			{MPCRst},
		},
		JEQ: {
			{LdFlags, PCOut, LdMAddr},
			{PCLd, PCLdIfEq, MEn, PCEnable},
			{MPCRst},
		},
		JGE: {
			{LdFlags, PCOut, LdMAddr},
			{PCLd, PCLdIfGorEq, MEn, PCEnable},
			{MPCRst},
		},
		PUSH: {
			{StkEn, LdMAddr},
			{MWr, EnA},
			{StkEn, StkLd, StkDec},
			{MPCRst},
		},
		POP: {
			{StkEn, StkInc, StkLd, LdMAddr},
			{LdA, MEn},
			{MPCRst},
		},
		CALL: {
			{StkEn, LdMAddr},
			{PCOut, MWr},
			{StkEn, StkLd, StkDec},
			{PCOut, LdMAddr},
			{PCLd, MEn, PCEnable},
			{MPCRst},
		},
		RET: {
			{StkEn, StkInc, StkLd, LdMAddr},
			{PCLd, MEn},
			{PCEnable},
			{MPCRst},
		},
		MOVa: {
			{LdA, EnB},
			{MPCRst},
		},
		MOVb: {
			{LdB, EnA},
			{MPCRst},
		},
		RST: {
			{EnGnd, PCLd, LdInst, LdA, LdB, LdOut, LdMAddr, StkRst},
			{RstDone},
		},
		OUT: {
			{EnA, LdOut},
			{MPCRst},
		},
		HLT: {
			{Hlt},
		},
	}

	buildRom(rom, preamble, instructions)

	writeRom(rom)
}

func buildRom(rom [][]ControlLine, preamble [][]ControlLine, instructions map[Instruction][][]ControlLine) {
	clockCycleLength := int(math.Pow(2, float64(pcBits)))
	for i := 0; i < len(rom)/clockCycleLength; i++ {
		for j, cyclePointer := (i * clockCycleLength), 0; cyclePointer < clockCycleLength; j, cyclePointer = j+1, cyclePointer+1 {
			if cyclePointer < len(preamble) {
				rom[j] = copyControlLines(preamble[cyclePointer])
				continue
			}
			if val, exists := instructions[Instruction(i)]; exists {
				if cyclePointer-len(preamble) < len(val) {
					rom[j] = copyControlLines(val[cyclePointer-len(preamble)])
				}
				// Else leave the control lines as 0, these control lines should never get hit
			}
			// Else we already wrote the preamble, and leave the rest of the control lines 0 since
			// These will be unreachable unless we get an invalid instruction
		}
	}
}

func writeRom(rom [][]ControlLine) {
	var sb strings.Builder
	sb.WriteString("v2.0 raw\n")
	f, err := os.OpenFile("rom.hex", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Error opening rom file %s\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()
	for i := 0; i < len(rom); i++ {
		sb.WriteString(outputMachineInstruction(rom[i]))
		sb.WriteString("\n")
	}
	f.WriteString(sb.String())
}

func outputMachineInstruction(activeControlLines []ControlLine) string {
	var sb strings.Builder
	linesMap := make(map[int]bool, 0)
	for i := 0; i < int(highestLine+1); i++ {
		linesMap[i] = false
	}
	for _, v := range activeControlLines {
		linesMap[int(v)] = true
	}
	// Reverse order because otherwise ix 0 of this loop
	// gets output as the highest bit in the result
	for i := int(highestLine); i >= 0; i-- {
		if linesMap[i] {
			sb.WriteString("1")
		} else {
			sb.WriteString("0")
		}
	}
	binaryValue := sb.String()
	fmt.Println(binaryValue)
	dataValue, _ := strconv.ParseInt(binaryValue, 2, 64)

	return fmt.Sprintf("%X", dataValue)
}

func copyControlLines(dst []ControlLine) []ControlLine {
	copyTarget := make([]ControlLine, len(dst))
	copy(copyTarget, dst)

	return copyTarget
}
