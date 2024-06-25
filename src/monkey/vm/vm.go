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
	stackPointer int
}

func New(bytecode *compiler.Bytecode) VM {
	return VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
	}
}

func (vm *VM) Execute() error {
	// Can we use this range or do we have to manually iterate?
	// I think we have to manually iterate because we have to be able to jump the instructionPointer
	for instructionPointer := 0; instructionPointer < len(vm.instructions); instructionPointer++ {
		// Fetch
		operation := opcode.OpCode(vm.instructions[instructionPointer])

		// Decode & Execute
		switch operation {
		case opcode.OpConstant:
			// Index of an OpConstant is two bytes wide
			// Don't look up width using opcode.Lookup, that is a lot of operations,
			// Hardcode that we know how big it is
			index := binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:])

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}

			instructionPointer += 2

		case opcode.OpAdd:
			left := vm.pop().(*object.Integer)
			right := vm.pop().(*object.Integer)

			result := object.Integer{
				Value: left.Value + right.Value,
			}

			err := vm.push(&result)
			if err != nil {
				return err
			}

		case opcode.OpPop:
			vm.pop()

		default:
			panic(fmt.Sprintf("Invalid opcode %d", operation))
		}
	}

	return nil
}

func (vm *VM) push(object object.Object) error {
	if vm.stackPointer >= len(vm.stack) {
		return fmt.Errorf("stack overflow (size %d)", cap(vm.stack))
	}

	vm.stack[vm.stackPointer] = object

	vm.stackPointer += 1

	return nil
}

func (vm *VM) pop() object.Object {
	result := vm.stack[vm.stackPointer-1]

	vm.stackPointer--

	return result
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		// Since object.Object is an interface,
		// we can return a nil value without changing the signature to *object.Object
		return nil
	}

	return vm.stack[vm.stackPointer-1]
}

// For tests
func (vm *VM) LastStackTop() object.Object {
	return vm.stack[vm.stackPointer]
}
