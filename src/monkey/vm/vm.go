package vm

import (
	"encoding/binary"
	"fmt"
	"monkey/compiler"
	"monkey/object"
	"monkey/opcode"
)

const StackSize = 2048

type VM struct {
	instructions opcode.Instructions
	constants    []object.Object
	stack        []object.Object
	// Next *free* slot in the stack, i.e. current length
	stack_pointer int
}

func New(bytecode *compiler.Bytecode) VM {
	return VM{
		instructions:  bytecode.Instructions,
		constants:     bytecode.Constants,
		stack:         make([]object.Object, StackSize),
		stack_pointer: 0,
	}
}

func (vm *VM) Execute() error {
	// Can we use this range or do we have to manually iterate?
	// I think we have to manually iterate because we have to be able to jump the instruction_pointer
	for instruction_pointer := 0; instruction_pointer < len(vm.instructions); instruction_pointer++ {
		// Fetch
		operation := opcode.OpCode(vm.instructions[instruction_pointer])

		// Decode & Execute
		switch operation {
		case opcode.OpConstant:
			// Index of an OpConstant is two bytes wide
			// Don't look up width using opcode.Lookup, that is a lot of operations,
			// Hardcode that we know how big it is
			index := binary.BigEndian.Uint16(vm.instructions[instruction_pointer+1:])

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}
			instruction_pointer += 2
		}
	}

	return nil
}

func (vm *VM) push(object object.Object) error {
	if vm.stack_pointer >= len(vm.stack) {
		return fmt.Errorf("stack overflow (size %d)", cap(vm.stack))
	}

	vm.stack[vm.stack_pointer] = object

	vm.stack_pointer += 1

	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.stack_pointer == 0 {
		// Since object.Object is an interface,
		// we can return a nil value without having to return a *object.Object
		return nil
	}

	return vm.stack[vm.stack_pointer-1]
}
