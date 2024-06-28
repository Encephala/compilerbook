package vm

import (
	"encoding/binary"
	"fmt"
	"monkey/compiler"
	"monkey/object"
	"monkey/opcode"
)

const StackSize = 2048
const GlobalsSize = 65536 // Matching sixteen-bit operand of OpSetGlobal/OpGetGlobal

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

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

	stack        [StackSize]object.Object
	stackPointer int // Next *free* slot in the stack, i.e. current length

	globals    *[GlobalsSize]object.Object
	numGlobals int
}

func New(bytecode *compiler.Bytecode) VM {
	return VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:        [StackSize]object.Object{nil},
		stackPointer: 0,

		globals:    &[GlobalsSize]object.Object{nil},
		numGlobals: 0,
	}
}

func NewWithState(bytecode *compiler.Bytecode, state *[GlobalsSize]object.Object) VM {
	return VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:        [StackSize]object.Object{},
		stackPointer: 0,

		globals:    state,
		numGlobals: 0,
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
		case opcode.OpGetConstant:
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

		case opcode.OpPushNull:
			err := vm.push(Null)
			if err != nil {
				return nil
			}

		case opcode.OpNegate:
			err := vm.executeNegate()
			if err != nil {
				return nil
			}

		case opcode.OpLogicalNot:
			err := vm.executeLogicalNot()
			if err != nil {
				return nil
			}

		case opcode.OpAdd, opcode.OpSubtract, opcode.OpMultiply, opcode.OpDivide,
			opcode.OpEquals, opcode.OpNotEquals, opcode.OpGreaterThan:
			err := vm.executeBinaryOperation(operation)

			if err != nil {
				return err
			}

		case opcode.OpJump:
			newPosition := int(binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:]))

			// Have to decrement by one because the instruction loop post-increments by one
			instructionPointer = newPosition - 1

		case opcode.OpJumpNotTruthy:
			condition := vm.pop()

			if !isTruthy(condition) {
				newPosition := int(binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:]))

				instructionPointer = newPosition - 1
			} else {
				// Skip jump target
				instructionPointer += 2
			}

		case opcode.OpSetGlobal:
			index := int(binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:]))

			vm.globals[index] = vm.pop()

			instructionPointer += 2

		case opcode.OpGetGlobal:
			index := int(binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:]))

			err := vm.push(vm.globals[index])
			if err != nil {
				return err
			}

			instructionPointer += 2

		case opcode.OpPop:
			vm.pop()

		case opcode.OpArray:
			length := int(binary.BigEndian.Uint16(vm.instructions[instructionPointer+1:]))

			result := &object.Array{}

			for i := range length {
				result.Elements = append(result.Elements, vm.stack[vm.stackPointer-length+i])
			}

			vm.stackPointer -= length

			err := vm.push(result)
			if err != nil {
				return err
			}

			instructionPointer += 2

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
	// This could underflow I guess
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

func (vm *VM) executeNegate() error {
	operand := vm.pop()

	value, ok := operand.(*object.Integer)

	if !ok {
		panic(fmt.Sprintf("Object %v not an integer", operand))
	}

	value.Value *= -1

	return vm.push(value)
}

func (vm *VM) executeLogicalNot() error {
	operand := vm.pop()

	result := toBoolObject(!isTruthy(operand))

	return vm.push(result)
}

func (vm *VM) executeBinaryOperation(operation opcode.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryOperationInteger(operation, left.(*object.Integer), right.(*object.Integer))
	}

	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return vm.executeBinaryOperationBoolean(operation, left.(*object.Boolean), right.(*object.Boolean))
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeBinaryOperationString(operation, left.(*object.String), right.(*object.String))
	}

	panic(fmt.Sprintf("Invalid operand types %T, %T", left, right))
}

func (vm *VM) executeBinaryOperationInteger(operation opcode.OpCode, left, right *object.Integer) error {
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

func (vm *VM) executeBinaryOperationBoolean(operation opcode.OpCode, left, right *object.Boolean) error {
	var result object.Object

	switch operation {
	case opcode.OpEquals:
		result = toBoolObject(left.Value == right.Value)

	case opcode.OpNotEquals:
		result = toBoolObject(left.Value != right.Value)

	default:
		panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
	}

	return vm.push(result)
}

func (vm *VM) executeBinaryOperationString(operation opcode.OpCode, left, right *object.String) error {
	var result object.Object

	switch operation {
	case opcode.OpAdd:
		result = &object.String{
			Value: left.Value + right.Value,
		}

	case opcode.OpEquals:
		result = toBoolObject(left.Value == right.Value)

	case opcode.OpNotEquals:
		result = toBoolObject(left.Value != right.Value)

	default:
		panic(fmt.Sprintf("Invalid opcode %q", opcode.Lookup(operation).Name))
	}

	return vm.push(result)
}

func isTruthy(value object.Object) bool {
	if value == True {
		return true
	}

	if value == False {
		return false
	}

	if value == Null {
		return false
	}

	integer, ok := value.(*object.Integer)
	if ok {
		// 0 is falsy, other integers are truthy
		// Deviating from the book here, which treats everything that isn't a boolean truthy
		return integer.Value != 0
	}

	panic(fmt.Sprintf("Object %v not booleanish", value))
}
