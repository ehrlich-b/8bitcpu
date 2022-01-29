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
	MPCRst
)

const (
	// Addressable ROM
	addressBits     = 8
	pcBits          = 3
	instructionBits = 5
	// Determines control width, update if new control lines are added
	highestLine = MPCRst
)

// [X, X, X, I, I, I, I, I]
// Leading 3 bits reserved for counter
const (
	NOP Instruction = 0b00000000
	LDA Instruction = 0b00000001
	ADD Instruction = 0b00000010
	JMP Instruction = 0b00000011
	STA Instruction = 0b00000100
	LDI Instruction = 0b00000101
	SUB Instruction = 0b00000110
	LDB Instruction = 0b00000111
	JEQ Instruction = 0b00001000
	JGE Instruction = 0b00001001
	JZ  Instruction = 0b00001010
	OUT Instruction = 0b00011110
	HLT Instruction = 0b00011111
)

// These machine instructions get executed for every command
var preamble = [][]ControlLine{
	{PCOut, LdMAddr},
	{MEn, LdInst, PCEnable},
}

func main() {
	rom := make([][]ControlLine, int(math.Pow(2, float64(addressBits))))
	// TODO Implement microcode PC reset
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
		ADD: {
			{PCOut, LdMAddr},
			{MEn, LdB},
			{ALUOut, LdA, PCEnable},
			{MPCRst},
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
