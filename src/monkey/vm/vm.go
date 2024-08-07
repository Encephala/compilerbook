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
const MaxFrames = 1024

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
	constants []object.Object

	stack        [StackSize]object.Object
	stackPointer int // Next *free* slot in the stack, i.e. current length

	globals    *[GlobalsSize]object.Object
	numGlobals int

	frames     [MaxFrames]*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) VM {
	mainFunction := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{
		Function:      mainFunction,
		FreeVariables: []object.Object{},
	}
	mainFrame := NewFrame(mainClosure, 0)

	return VM{
		constants: bytecode.Constants,

		stack:        [StackSize]object.Object{nil},
		stackPointer: 0,

		globals:    &[GlobalsSize]object.Object{nil},
		numGlobals: 0,

		frames:     [MaxFrames]*Frame{0: mainFrame},
		frameIndex: 0,
	}
}

func NewWithState(bytecode *compiler.Bytecode, state *[GlobalsSize]object.Object) VM {
	mainFunction := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{
		Function:      mainFunction,
		FreeVariables: []object.Object{},
	}
	mainFrame := NewFrame(mainClosure, 0)

	return VM{
		constants: bytecode.Constants,

		stack:        [StackSize]object.Object{},
		stackPointer: 0,

		globals:    state,
		numGlobals: 0,

		frames:     [MaxFrames]*Frame{0: mainFrame},
		frameIndex: 0,
	}
}

func (vm *VM) Execute() error {
	var instructionPointer int
	var instructions opcode.Instructions
	var operation opcode.OpCode

	for vm.currentFrame().instructionPointer < len(*vm.currentFrame().Instructions()) {
		instructionPointer = vm.currentFrame().instructionPointer
		instructions = *vm.currentFrame().Instructions()

		// Determine current instruction, then increment
		vm.currentFrame().instructionPointer++

		// Fetch
		operation = opcode.OpCode(instructions[instructionPointer])

		// Decode & Execute
		switch operation {
		case opcode.OpGetConstant:
			// Index of an OpConstant is two bytes wide
			// Don't look up width using opcode.Lookup, that is a lot of operations,
			// Hardcode that we know how big it is
			index := binary.BigEndian.Uint16(instructions[instructionPointer+1:])

			err := vm.push(vm.constants[index])
			if err != nil {
				return err
			}

			vm.currentFrame().instructionPointer += 2

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
			newPosition := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

			vm.currentFrame().instructionPointer = newPosition

		case opcode.OpJumpNotTruthy:
			condition := vm.pop()

			if !isTruthy(condition) {
				newPosition := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

				vm.currentFrame().instructionPointer = newPosition
			} else {
				// Skip jump target
				vm.currentFrame().instructionPointer += 2
			}

		case opcode.OpSetGlobal:
			index := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

			vm.globals[index] = vm.pop()

			vm.currentFrame().instructionPointer += 2

		case opcode.OpGetGlobal:
			index := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

			err := vm.push(vm.globals[index])
			if err != nil {
				return err
			}

			vm.currentFrame().instructionPointer += 2

		case opcode.OpPop:
			vm.pop()

		case opcode.OpArray:
			length := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

			result := &object.Array{}

			for i := range length {
				result.Elements = append(result.Elements, vm.stack[vm.stackPointer-length+i])
			}

			vm.stackPointer -= length

			err := vm.push(result)
			if err != nil {
				return err
			}

			vm.currentFrame().instructionPointer += 2

		case opcode.OpHash:
			length := int(binary.BigEndian.Uint16(instructions[instructionPointer+1:]))

			result := &object.Hash{Pairs: make(map[object.HashKey]object.HashPair)}

			for i := range length {
				key := vm.stack[vm.stackPointer-length*2+2*i]
				value := vm.stack[vm.stackPointer-length*2+2*i+1]

				hashKey, ok := key.(object.Hashable)
				if !ok {
					return fmt.Errorf("INVALID HASH KEY: %s", key.Type())
				}

				result.Pairs[hashKey.HashKey()] = object.HashPair{
					Key:   key,
					Value: value,
				}
			}

			vm.stackPointer -= length * 2

			err := vm.push(result)
			if err != nil {
				return err
			}

			vm.currentFrame().instructionPointer += 2

		case opcode.OpIndex:
			index := vm.pop()
			indexee := vm.pop()

			err := vm.executeIndexExpression(indexee, index)
			if err != nil {
				return err
			}

		case opcode.OpCall:
			numberOfArguments := int(instructions[instructionPointer+1])
			vm.currentFrame().instructionPointer++

			basePointer := vm.stackPointer - numberOfArguments
			function := vm.stack[basePointer-1]

			switch callee := function.(type) {
			case *object.Closure:
				if numberOfArguments != callee.Function.NumberOfParameters {
					return fmt.Errorf("wrong number of arguments %d, expected %d", numberOfArguments, callee.Function.NumberOfParameters)
				}

				frame := NewFrame(callee, basePointer)
				vm.pushFrame(frame)

				vm.stackPointer += callee.Function.NumberOfLocals

			case *object.Builtin:
				arguments := vm.stack[basePointer:vm.stackPointer]

				result := callee.Fn(arguments...)
				vm.stackPointer = basePointer - 1

				if result == nil {
					vm.push(Null)
				} else {
					vm.push(result)
				}

			default:
				return fmt.Errorf("TRIED CALLING NON-FUNCTION")
			}

		case opcode.OpSetLocal:
			index := int(instructions[instructionPointer+1])
			vm.currentFrame().instructionPointer += 1

			value := vm.pop()

			vm.stack[vm.currentFrame().basePointer+index] = value

		case opcode.OpGetLocal:
			index := int(instructions[instructionPointer+1])
			vm.currentFrame().instructionPointer += 1

			value := vm.stack[vm.currentFrame().basePointer+index]

			err := vm.push(value)
			if err != nil {
				return err
			}

		case opcode.OpReturnValue:
			frame := vm.popFrame()

			returnValue := vm.pop()

			vm.stackPointer = frame.basePointer

			vm.stack[vm.stackPointer-1] = returnValue

		case opcode.OpReturn:
			frame := vm.popFrame()

			vm.stackPointer = frame.basePointer

			vm.stack[vm.stackPointer-1] = Null

		case opcode.OpGetBuiltin:
			index := int(instructions[instructionPointer+1])
			vm.currentFrame().instructionPointer++

			definition := object.Builtins[index]

			err := vm.push(definition.Builtin)
			if err != nil {
				return nil
			}

		case opcode.OpMakeClosure:
			index := binary.BigEndian.Uint16(instructions[instructionPointer+1:])
			numberOfFreeVariables := int(instructions[instructionPointer+3])
			vm.currentFrame().instructionPointer += 3

			err := vm.pushClosure(int(index), numberOfFreeVariables)
			if err != nil {
				return err
			}

		case opcode.OpGetFree:
			index := int(instructions[instructionPointer+1])
			vm.currentFrame().instructionPointer++

			variable := vm.currentFrame().closure.FreeVariables[index]

			err := vm.push(variable)
			if err != nil {
				return err
			}

		case opcode.OpRecurse:
			currentClosure := vm.currentFrame().closure
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}

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

	vm.stackPointer++

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

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex]
}

func (vm *VM) pushFrame(frame *Frame) {
	vm.frameIndex++
	vm.frames[vm.frameIndex] = frame
}

func (vm *VM) popFrame() *Frame {
	result := vm.frames[vm.frameIndex]

	vm.frameIndex--
	return result
}

func (vm *VM) pushClosure(index int, numberOfFreeVariables int) error {
	constant := vm.constants[index]

	converted, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("NOT A FUNCTION: %+v", constant)
	}

	freeVariables := make([]object.Object, numberOfFreeVariables)
	copy(freeVariables, vm.stack[vm.stackPointer-numberOfFreeVariables:vm.stackPointer])
	vm.stackPointer -= numberOfFreeVariables

	closure := &object.Closure{
		Function:      converted,
		FreeVariables: freeVariables,
	}

	return vm.push(closure)
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

func (vm *VM) executeIndexExpression(indexee, index object.Object) error {
	switch indexee := indexee.(type) {
	case *object.Array:
		convertedIndex, ok := index.(*object.Integer)
		if !ok {
			return fmt.Errorf("INVALID ARRAY INDEX: %v", index)
		}

		if convertedIndex.Value < 0 || convertedIndex.Value >= int64(len(indexee.Elements)) {
			return vm.push(Null)
		}

		return vm.push(indexee.Elements[convertedIndex.Value])

	case *object.Hash:
		convertedIndex, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("INVALID HASH INDEX: %v", index)
		}

		result, ok := indexee.Pairs[convertedIndex.HashKey()]

		if !ok {
			return vm.push(Null)
		}

		return vm.push(result.Value)

	default:
		panic(fmt.Sprintf("unexpected object.Object: %#v", indexee))
	}
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
