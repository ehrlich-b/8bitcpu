
This is an 8 bit CPU, loosely based on [Ben Eater's](https://eater.net/8bit) 8 bit computer. This CPU is meant to be run in the [Digital](https://github.com/hneemann/Digital) circuit simulator. This repository also contains two small go programs, one to generate the microcode rom, and one assembler that writes directly to the instruction rom.

![](https://raw.githubusercontent.com/adotout/8bitcpu/main/exports/CPU.svg)

# Specs

* Up to 5 kHz clock speed
* 32 bytes of RAM
* One output register with an 8 bit hex display
* Reset button
* Two general purpose registers
* Basic stack features
* Instruction ROM => memory loader, so you don't have to hand program the CPU every boot

# Programming the CPU

This CPU use an extra compact instruction set, all instructions have either 0 or 1 parameters.

## Syntax

Lines containing only whitespace are ignored.

Each non-whitespace line is either a label, or an instruction, with an optional parameter:

Instructions
```
[Instruction name] [optional parameter]
```

Parameters can be literal values, or label names.

Examples:
```
LDAi $1
LDBi $1
JEQ labelname
ADD
```

Literal parameter values are 8 bit numbers starting with $
```
$[literal number value]
```
Examples:
```
$10    ## Decimal "10"
$0x10  ## Decimal "16"
$0b11  ## Decimal "3"
```

Labels are arbitrary strings followed by ":"
```
[label_text]:
```

Examples:
```
label:
loop:
```

## Instructions

### Load instructions LDA, LDB, LDAi, LDBi

`LD[Register][immediate?] [required parameter]`

Load value into register `[Register]`. If `i` "immediate", interpret the provided value as a literal number, otherwise interpret the number as an address, and load the value from that address.

Examples:
```
LDAi $10 ## Load the value "10" into register "A"
LDB $4   ## Load the value at memory address "4" into register "B"
```

### STA

`STA [required parameter]`

Store the value in A at memory address `[parameter]`.

### Arithmetic ADD, SUB

Add or subtract A [+|-] B, and store the result in A.

### Jump instructions JMP, JZ, JEQ, JGE

`JMP [required label]`

Jump unconditionally to `[label]`.

`JZ [required label]`

Jump to `[label]` if Register A = 0.

`JEQ [required label]`

Jump to `[label]` if Register A = Register B.

`JGE [required label]`

Jump to `[label]` if Register A >= Register B.

### Output OUT

Display the value in register A on the hex display.

### Move instructions MOVa, MOVa

`MOV[register name]`

`MOVa` - move B into A
`MOVb` - move A into B

### Halt HLT

Halt the computer.

### Stack instructions PUSH, POP, CALL, RET

`PUSH`

Push the value in A onto the stack.

`POP`

Pop the value off the top of the stack, and store the result in A.

`CALL [required label]`

Store the program counter on the stack, and jump to `[label]`.

`RET`

Pop the top of the stack, and jump to the address (address can be set to the PC with "CALL", but this isn't strictly required).

# Local setup

This is going to be a little cumbersome. If anyone ever reads this: contributions welcome :)

We are going to install Digital, and then checkout this project into Digital's installation location, so that you'll have easy access to the custom components of this project, directly in your Digital installation.

* [Install Digital](https://github.com/hneemann/Digital/releases) by unzipping their release into a folder.
* cd into to the install location and run
  * `git init`
  * `git remote add origin git@github.com:adotout/8bitcpu.git`
  * `git fetch`
  * `git checkout origin/main -ft`
* At this point you have installed Digital, and pulled in the custom components required for the CPU.

The CPU and control_logic components need to reference ROM components on your computer. Digital does not appear to support relative paths, so you'll have to:

* Copy `custom_components/CPU.dig.dist` to `custom_components/CPU.dig`
* Copy `custom_components/control_logic.dig.dist` to `custom_components/control_logic.dig`
* Replace the instances of `{{Your Digital installation location}}` in both files, with your Digital installation path.
  * Example: `<file>{{Your Digital installation location}}\microcode\rom.hex</file>` => `<file>C:\Users\me\Digital\microcode\rom.hex</file>`

You should be all set, open the Digital .jar or .exe file, then open `custom_components\CPU.dig`, and run the default program (at the time of writing is a program that counts from 1 => 10).
