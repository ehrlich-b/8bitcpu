package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Instruction int8

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
	OUT  Instruction = 0b00011110 // No argument, display the value stored in register A
	HLT  Instruction = 0b00011111 // No argument, halt the CPU
)

// This seems less than ideal, but golang doesn't have real enums
var instructionMap = map[string]Instruction{
	"NOP":  NOP,
	"LDA":  LDA,
	"LDAi": LDAi,
	"LDB":  LDB,
	"LDBi": LDBi,
	"STA":  STA,
	"ADD":  ADD,
	"SUB":  SUB,
	"JMP":  JMP,
	"JZ":   JZ,
	"JEQ":  JEQ,
	"JGE":  JGE,
	"OUT":  OUT,
	"HLT":  HLT,
	"MOVa": MOVa,
	"MOVb": MOVb,
	"PUSH": PUSH,
	"POP":  POP,
	"CALL": CALL,
	"RET":  RET,
}

func main() {
	//instructions := make([]int8, 0)
	args := os.Args
	var program string
	if len(args) < 2 {
		program = "./programs/callret.asm"
	} else {
		program = os.Args[2]
	}
	instructions, err := loadProgram(program)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = writeRom(instructions)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func loadProgram(assemblyFile string) ([]Instruction, error) {
	b, err := ioutil.ReadFile(assemblyFile) // just pass the file name
	if err != nil {
		return nil, fmt.Errorf("Error reading assembly file %s: %s\n", assemblyFile, err.Error())
	}
	rawContent := string(b)
	content := cleanInput(rawContent)
	labels, err := setupLabels(content)
	if err != nil {
		return nil, err
	}

	return writeInstructions(content, labels)
}

func writeInstructions(program [][]string, labels map[string]int8) ([]Instruction, error) {
	var memoryCounter int8
	instructions := make([]Instruction, 0)
nextLine:
	for lineIx, progLine := range program {
		if len(progLine) == 0 {
			continue
		}
		for instrIx, instr := range progLine {
			// Skip labels
			if strings.HasSuffix(instr, ":") {
				continue nextLine
			}
			// This is an instruction argument
			if instrIx != 0 {
				if strings.HasPrefix(instr, "$") { // Parse the argument as a literal value
					number := strings.TrimLeft(instr, "$")
					argument, err := strconv.ParseInt(number, 0, 9)
					if err != nil {
						return nil, fmt.Errorf("Unable to parse argument %s as int on line %d", instr, lineIx)
					}
					instructions = append(instructions, Instruction(argument))
				} else { // Otherwise it must be a label
					if addr, exists := labels[instr]; exists {
						instructions = append(instructions, Instruction(addr))
					} else {
						return nil, fmt.Errorf("Undefined label '%s' on line %d", instr, lineIx)
					}
				}
				continue
			}
			if instruction, exists := instructionMap[instr]; exists {
				instructions = append(instructions, Instruction(instruction))
			} else {
				return nil, fmt.Errorf("Undefined instruction '%s' on line %d", instr, lineIx)
			}
			memoryCounter++
		}
	}
	return instructions, nil
}

func setupLabels(program [][]string) (map[string]int8, error) {
	labels := make(map[string]int8, 0)
	var memoryCounter int8
	for lineIx, progLine := range program {
		if len(progLine) == 0 {
			continue
		}
		for _, instr := range progLine {
			if strings.HasSuffix(instr, ":") {
				label := strings.TrimRight(instr, ":")
				if _, exists := labels[label]; exists {
					return nil, fmt.Errorf("Duplicate label %s at line %d", label, lineIx)
				}
				labels[label] = memoryCounter
			} else {
				memoryCounter++
			}
		}
	}

	return labels, nil
}

func cleanInput(rawContent string) [][]string {
	rawLines := strings.Split(rawContent, "\n")
	cleanInput := make([][]string, 0)
	for _, line := range rawLines {
		line = strings.Split(line, "//")[0]
		tokens := strings.Fields(strings.TrimSpace(line))
		cleanInput = append(cleanInput, tokens)
	}
	return cleanInput
}

func writeRom(prog []Instruction) error {
	var sb strings.Builder
	sb.WriteString("v2.0 raw\n")
	f, err := os.OpenFile("prog.hex", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("Error opening rom file %s\n", err.Error())
	}
	defer f.Close()
	for i := 0; i < len(prog); i++ {
		sb.WriteString(fmt.Sprintf("%X", uint8(prog[i])))
		sb.WriteString("\n")
	}
	f.WriteString(sb.String())

	return nil
}
