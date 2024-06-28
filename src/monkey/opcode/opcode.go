package opcode

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instruction []byte
type Instructions []byte

func (instructions Instructions) String() string {
	var out bytes.Buffer

	offset := 0
	for offset < len(instructions) {
		definition := Lookup(OpCode(instructions[offset]))

		operands, read := ReadOperands(definition, instructions[offset+1:])

		fmt.Fprintf(
			&out, "%04d %s\n",
			offset, fmtInstruction(definition, operands),
		)

		offset += 1 + int(read)
	}

	return out.String()
}

func fmtInstruction(definition *OpDefinition, operands []int) string {
	switch len(operands) {
	case 0:
		return definition.Name
	case 1:
		return fmt.Sprintf("%s %d", definition.Name, operands[0])

	default:
		panic(fmt.Sprintf("Invalid operand count %d for %s\n", len(operands), definition.Name))
	}
}

type OpCode byte

const (
	OpGetConstant OpCode = iota

	OpNegate
	OpLogicalNot

	OpAdd
	OpSubtract
	OpMultiply
	OpDivide

	OpEquals
	OpNotEquals
	OpGreaterThan

	OpPushTrue
	OpPushFalse
	OpPushNull

	OpJump
	OpJumpNotTruthy

	OpGetGlobal
	OpSetGlobal

	OpPop

	OpArray
)

type OpDefinition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*OpDefinition{
	OpGetConstant: {"OpGetConstant", []int{2}}, // Takes two bytes, so up to 65536 constants may be defined

	OpNegate:     {"OpNegate", []int{}},
	OpLogicalNot: {"OpLogicalNot", []int{}},

	OpAdd:      {"OpAdd", []int{}},
	OpSubtract: {"OpSubtract", []int{}},
	OpMultiply: {"OpMultiply", []int{}},
	OpDivide:   {"OpDivide", []int{}},

	OpEquals:      {"OpEquals", []int{}},
	OpNotEquals:   {"OpNotEquals", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},

	OpPushTrue:  {"OpPushTrue", []int{}},
	OpPushFalse: {"OpPushFalse", []int{}},
	OpPushNull:  {"OpPushNull", []int{}},

	OpJump:          {"OpJump", []int{2}}, // Program can be up to 65536 instructions long
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetGlobal: {"OpSetGlobal", []int{2}},

	OpPop: {"OpPop", []int{}},

	OpArray: {"OpArray", []int{2}},
}

// Book passes a byte as code, I pass the OpCode
func Lookup(code OpCode) *OpDefinition {
	result, ok := definitions[code]
	if !ok {
		panic(fmt.Sprintf("Opcode %d has not been defined", code))
	}

	return result
}

func MakeInstruction(code OpCode, operands ...int) Instruction {
	definition := Lookup(code)

	instructionsLength := 1
	for _, length := range definition.OperandWidths {
		instructionsLength += length
	}

	result := make(Instruction, instructionsLength)

	result[0] = byte(code)

	offset := 1
	for i, operand := range operands {
		switch definition.OperandWidths[i] {
		case 1:
			result[offset] = uint8(operand)

		case 2:
			binary.BigEndian.PutUint16(result[offset:], uint16(operand))

		default:
			panic(fmt.Sprintf("Invalid operand width: %d", definition.OperandWidths[i]))
		}

		offset += definition.OperandWidths[i]
	}

	return result
}

func ReadOperands(definition *OpDefinition, rawOperands []byte) ([]int, int) {
	operands := make([]int, len(definition.OperandWidths))

	offset := 0
	for i, width := range definition.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(rawOperands[offset])

		case 2:
			operands[i] = int(binary.BigEndian.Uint16(rawOperands[offset:]))

		default:
			panic(fmt.Sprintf("Invalid operand width %d", width))
		}

		offset += width
	}

	return operands, offset
}
