package intcode

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Memory = []int

type ParameterMode = int

var ErrSegFault = errors.New("SEGFAULT")
var ErrHalt = errors.New("HALT")
var ReadError = errors.New("READ")

const (
	IMEDIATE_MODE ParameterMode = 1
	RELATIVE_MODE ParameterMode = 2
	POSITION_MODE ParameterMode = 0
)

type Parameter struct {
	Mode ParameterMode
	Arg  int
}

func (p Parameter) GetByMode(vm *IntCodeVM) (int, error) {
	switch p.Mode {
	case IMEDIATE_MODE:
		return p.Arg, nil
	case RELATIVE_MODE:
		offset := vm.RelativeBase + p.Arg
		if offset >= len(vm.Memory) || offset < 0 {
			return 0, ErrSegFault
		}
		return vm.Memory[offset], nil
	default:
		if p.Arg >= len(vm.Memory) {
			return 0, ErrSegFault
		}
		return vm.Memory[p.Arg], nil
	}
}

func (p Parameter) SaveByMode(vm *IntCodeVM, value int) error {
	var offset int
	switch p.Mode {
	case RELATIVE_MODE:
		offset = vm.RelativeBase + p.Arg
	case IMEDIATE_MODE:
		return errors.New("IMEDIATE_MODE memset")
	default:
		offset = p.Arg
	}
	if offset < 0 || offset >= len(vm.Memory) {
		return ErrSegFault
	}
	vm.Memory[offset] = value
	return nil
}

type OpCode struct {
	Mnemonic  string
	Number    int
	Argc      int
	Operation func(*IntCodeVM, []Parameter) error
}

type IntCodeIO interface {
	Read() (int, error)
	Write(int)
}

type IntCodePipe struct {
	Buffer []int
}

func (pipe *IntCodePipe) Read() (int, error) {
	var x int
	if len(pipe.Buffer) < 1 {
		return 0, ReadError
	}
	x, pipe.Buffer = pipe.Buffer[0], pipe.Buffer[1:]
	return x, nil
}

func (pipe *IntCodePipe) Write(n int) {
	pipe.Buffer = append(pipe.Buffer, n)
}

type IntCodeVM struct {
	OpCodes           map[int]OpCode
	IntructionPointer int
	RelativeBase      int

	Memory Memory

	Stdin  IntCodeIO
	Stdout IntCodeIO
}

func ParseParameterModes(instruction int) (map[int]ParameterMode, int) {
	modes := make(map[int]ParameterMode)
	opcode := instruction % 100
	modebits := instruction / 100
	index := 0
	for modebits != 0 {
		modes[index] = modebits % 10
		modebits = modebits / 10
		index += 1
	}
	//log.Printf("\ninstruction=%d\nmodes=%v\nopcode=%d", instruction, modes, opcode)
	return modes, opcode
}

func (vm *IntCodeVM) Step() error {
	instruction, err := vm.GetInstruction()
	if err != nil {
		return err
	}
	oldIp := vm.IntructionPointer
	modes, op := ParseParameterModes(instruction)
	opcode, ok := vm.OpCodes[op]
	if !ok {
		return errors.New(fmt.Sprintf("Unknown Opcode %d", op))
	}
	parms := make([]Parameter, opcode.Argc)
	var offset int
	for i, _ := range parms {
		offset = vm.IntructionPointer + i + 1
		if offset > len(vm.Memory) {
			return ErrSegFault
		}
		mode, ok := modes[i]
		if !ok {
			mode = POSITION_MODE
		}
		parms[i] = Parameter{
			Mode: mode,
			Arg:  vm.Memory[offset],
		}
	}
	//log.Printf("%s(%v)\n", opcode.Mnemonic, parms)
	//log.Printf("%d\n%v", vm.IntructionPointer, vm.Memory)
	err = opcode.Operation(vm, parms)
	if err != nil {
		return err
	}

	if oldIp == vm.IntructionPointer {
		vm.IntructionPointer += (1 + opcode.Argc)
	}

	// Everything is fine
	return nil
}

func (vm *IntCodeVM) RunUntilError() (err error) {
	for {
		err = vm.Step()
		if err != nil {
			return
		}
	}
}

func (vm *IntCodeVM) RunUntilOutput() (output int, err error) {
	for {
		err = vm.Step()
		if err != nil {
			return
		}
		output, err = vm.Stdout.Read()
		switch err {
		case ReadError:
		default: // should be both nonReadErrors and output
			return
		}
	}
}

func (vm *IntCodeVM) RunUntilHalt() error {
	err := vm.RunUntilError()
	if err != ErrHalt {
		return err
	}
	return nil
}

func (vm *IntCodeVM) GetInstruction() (int, error) {
	if vm.IntructionPointer > len(vm.Memory) {
		return 0, ErrSegFault
	}
	return vm.Memory[vm.IntructionPointer], nil
}

func NewVM(program string) *IntCodeVM {
	vm := new(IntCodeVM)
	vm.OpCodes = make(map[int]OpCode)
	for _, opcode := range OpCodes {
		vm.OpCodes[opcode.Number] = opcode
	}
	programCode, err := ParseProgram(program)
	if err != nil {
		// Do Something
	}
	memory := make([]int, 1024*100) // 100K should be good enough
	for i, instruction := range programCode {
		memory[i] = instruction
	}
	vm.Memory = memory

	vm.Stdin = new(IntCodePipe)
	vm.Stdout = new(IntCodePipe)

	return vm
}

func ParseProgram(program string) (Memory, error) {
	items := strings.Split(program, ",")
	memory := make(Memory, len(items))
	for i, item := range items {
		value, err := strconv.Atoi(item)
		if err != nil {
			return memory, err
		}
		memory[i] = value
	}
	return memory, nil
}

var OpCodes = [...]OpCode{
	OpCode{"add", 1, 3, func(vm *IntCodeVM, paramters []Parameter) error {
		a, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		b, err := paramters[1].GetByMode(vm)
		if err != nil {
			return err
		}
		err = paramters[2].SaveByMode(vm, a+b)
		if err != nil {
			return err
		}
		return nil
	}},
	OpCode{"mult", 2, 3, func(vm *IntCodeVM, paramters []Parameter) error {
		a, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		b, err := paramters[1].GetByMode(vm)
		if err != nil {
			return err
		}
		err = paramters[2].SaveByMode(vm, a*b)
		if err != nil {
			return err
		}
		return nil
	}},
	OpCode{"input", 3, 1, func(vm *IntCodeVM, paramters []Parameter) error {
		value, err := vm.Stdin.Read()
		if err != nil {
			return err
		}
		err = paramters[0].SaveByMode(vm, value)
		if err != nil {
			return err
		}
		return nil
	}},
	OpCode{"output", 4, 1, func(vm *IntCodeVM, paramters []Parameter) error {
		value, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		vm.Stdout.Write(value)
		return nil
	}},
	OpCode{"jmp-if-true", 5, 2, func(vm *IntCodeVM, paramters []Parameter) error {
		value, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		if value != 0 {
			location, err := paramters[1].GetByMode(vm)
			if err != nil {
				return err
			}
			vm.IntructionPointer = location
		}
		return nil
	}},
	OpCode{"jmp-if-false", 6, 2, func(vm *IntCodeVM, paramters []Parameter) error {
		value, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		if value == 0 {
			location, err := paramters[1].GetByMode(vm)
			if err != nil {
				return err
			}
			vm.IntructionPointer = location
		}
		return nil
	}},
	OpCode{"less-than", 7, 3, func(vm *IntCodeVM, paramters []Parameter) error {
		a, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		b, err := paramters[1].GetByMode(vm)
		if err != nil {
			return err
		}
		if a < b {
			err := paramters[2].SaveByMode(vm, 1)
			if err != nil {
				return err
			}
		} else {
			err := paramters[2].SaveByMode(vm, 0)
			if err != nil {
				return err
			}
		}
		return nil
	}},
	OpCode{"equals", 8, 3, func(vm *IntCodeVM, paramters []Parameter) error {
		a, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		b, err := paramters[1].GetByMode(vm)
		if err != nil {
			return err
		}
		if a == b {
			err := paramters[2].SaveByMode(vm, 1)
			if err != nil {
				return err
			}
		} else {
			err := paramters[2].SaveByMode(vm, 0)
			if err != nil {
				return err
			}
		}
		return nil
	}},
	OpCode{"adjust-base", 9, 1, func(vm *IntCodeVM, paramters []Parameter) error {
		base, err := paramters[0].GetByMode(vm)
		if err != nil {
			return err
		}
		vm.RelativeBase += base
		return nil
	}},
	OpCode{"halt", 99, 0, func(vm *IntCodeVM, paramters []Parameter) error {
		return ErrHalt
	}},
}
