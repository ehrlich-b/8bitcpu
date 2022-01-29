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
}

// Convert these to ints if there's a desire
// to support more than one argument
var arguments = map[Instruction]bool{
	LDA:  true,
	LDAi: true,
	LDB:  true,
	LDBi: true,
	STA:  true,
	JMP:  true,
	JZ:   true,
	JEQ:  true,
	JGE:  true,
}

func main() {
	//instructions := make([]int8, 0)
	args := os.Args
	var program string
	if len(args) < 2 {
		program = "./programs/fib.asm"
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

func writeInstructions(program []string, labels map[string]int8) ([]Instruction, error) {
	var memoryCounter int8
	argNext := false
	instructions := make([]Instruction, 0)
	for lineIx, progLine := range program {
		// Skip labels
		if strings.HasSuffix(progLine, ":") {
			continue
		}
		// This is an instruction argument
		if argNext {
			if strings.HasPrefix(progLine, "$") { // Parse the argument as a literal value
				number := strings.TrimLeft(progLine, "$")
				argument, err := strconv.ParseInt(number, 0, 9)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse argument %s as int on line %d", progLine, lineIx)
				}
				instructions = append(instructions, Instruction(argument))
			} else { // Otherwise it must be a label
				if addr, exists := labels[progLine]; exists {
					instructions = append(instructions, Instruction(addr))
				} else {
					return nil, fmt.Errorf("Undefined label '%s' on line %d", progLine, lineIx)
				}
			}
			argNext = false
			continue
		}
		if instruction, exists := instructionMap[progLine]; exists {
			instructions = append(instructions, Instruction(instruction))
			if val, exists := arguments[instruction]; val && exists {
				argNext = true
			}
		} else {
			return nil, fmt.Errorf("Undefined instruction '%s' on line %d", progLine, lineIx)
		}
		memoryCounter++
	}
	return instructions, nil
}

func setupLabels(program []string) (map[string]int8, error) {
	labels := make(map[string]int8, 0)
	var memoryCounter int8
	for lineIx, progLine := range program {
		if strings.HasSuffix(progLine, ":") {
			label := strings.TrimRight(progLine, ":")
			if _, exists := labels[label]; exists {
				return nil, fmt.Errorf("Duplicate label %s at line %d", label, lineIx)
			}
			labels[label] = memoryCounter
		} else {
			memoryCounter++
		}
	}

	return labels, nil
}

func cleanInput(rawContent string) []string {
	rawLines := strings.Split(rawContent, "\n")
	cleanInput := make([]string, 0)
	for _, line := range rawLines {
		tokens := strings.Fields(strings.TrimSpace(line))
		cleanInput = append(cleanInput, tokens...)
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
