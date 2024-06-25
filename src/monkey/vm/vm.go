package vm

import (
	"encoding/binary"
	"fmt"
	"monkey/compiler"
	"monkey/object"
	"monkey/opcode"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

func toBoolObject(b bool) *object.Boolean {
	if b {
		return True
	} else {
		return False
	}
}

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
		case opcode.OpReadConstant:
			// Index of an OpConstant is two bytes wide
			// Don't look up width using opcode.Lookup, that is a lot of operations,
			// Hardcode that we know how big it is
			index := binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:])

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}

			instructionPointer += 2

		case opcode.OpPushTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case opcode.OpPushFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case opcode.OpAdd, opcode.OpSubtract, opcode.OpMultiply, opcode.OpDivide:
			err := vm.executeBinaryArithmetic(operation)

			if err != nil {
				return err
			}

		case opcode.OpEquals, opcode.OpNotEquals, opcode.OpGreaterThan:
			err := vm.executeComparison(operation)

			if err != nil {
				return nil
			}

		case opcode.OpPop:
			vm.pop()

		default:
			panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
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

func (vm *VM) executeBinaryArithmetic(operation opcode.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryArithmeticInteger(operation, left.(*object.Integer), right.(*object.Integer))
	}

	panic(fmt.Sprintf("Invalid operand types %T, %T", left.Type(), right.Type()))
}

func (vm *VM) executeBinaryArithmeticInteger(operation opcode.OpCode, left, right *object.Integer) error {
	var result object.Object

	switch operation {
	case opcode.OpAdd:
		result = &object.Integer{
			Value: left.Value + right.Value,
		}

	case opcode.OpSubtract:
		result = &object.Integer{
			Value: left.Value - right.Value,
		}

	case opcode.OpMultiply:
		result = &object.Integer{
			Value: left.Value * right.Value,
		}

	case opcode.OpDivide:
		result = &object.Integer{
			Value: left.Value / right.Value,
		}

	case opcode.OpEquals:
		result = toBoolObject(left.Value == right.Value)

	case opcode.OpNotEquals:
		result = toBoolObject(left.Value != right.Value)

	case opcode.OpGreaterThan:
		result = toBoolObject(left.Value > right.Value)

	default:
		panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
	}

	return vm.push(result)
}

func (vm *VM) executeComparison(operation opcode.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeComparisonInteger(operation, left.(*object.Integer), right.(*object.Integer))
	}

	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return vm.executeComparisonBoolean(operation, left, right)
	}

	panic(fmt.Sprintf("Invalid operand types %T, %T", left.Type(), right.Type()))
}

func (vm *VM) executeComparisonInteger(operation opcode.OpCode, left, right *object.Integer) error {
	var result object.Object

	// Pointer comparison, True and False are global (semantically constant) Boolean objects
	switch operation {
	case opcode.OpEquals:
		result = toBoolObject(left.Value == right.Value)

	case opcode.OpNotEquals:
		result = toBoolObject(left.Value != right.Value)

	case opcode.OpGreaterThan:
		result = toBoolObject(left.Value > right.Value)

	default:
		panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
	}

	return vm.push(result)
}

func (vm *VM) executeComparisonBoolean(operation opcode.OpCode, left, right object.Object) error {
	var result object.Object

	// Pointer comparison, True and False are global (semantically constant) Boolean objects
	switch operation {
	case opcode.OpEquals:
		result = toBoolObject(left == right)

	case opcode.OpNotEquals:
		result = toBoolObject(left != right)

	default:
		panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
	}

	return vm.push(result)
}
