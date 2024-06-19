package opcode

import (
	"encoding/binary"
	"fmt"
)

type Instruction []byte
type Instructions []byte

type OpCode byte

const (
	OpConstant OpCode = iota
)

type OpDefinition struct {
	Name          string
	OperandWidths []uint
}

var definitions = map[OpCode]*OpDefinition{
	// Takes two bytes, so up to 65536 constants may be defined
	OpConstant: {"OpConstant", []uint{2}},
}

func Lookup(code OpCode) (*OpDefinition, error) {
	var e error = nil

	result, ok := definitions[code]
	if !ok {
		e = fmt.Errorf("opcode %d undefined", code)
	}

	return result, e
}

func Make(code OpCode, operands ...uint) Instruction {
	definition, err := Lookup(code)

	if err != nil {
		// There's probably a better way to handle constant not existing error
		// than returning empty, but whatever?
		// To be fair (as book points out)
		// Only matters for testing/debugging, as we're the only ones actually calling this function
		// Tests will catch this faulty behaviour
		return []byte{}
	}

	var instructionsLength uint = 1
	for _, length := range definition.OperandWidths {
		instructionsLength += length
	}

	result := make(Instruction, instructionsLength)

	result[0] = byte(code)

	var offset uint = 1
	for i, operand := range operands {
		switch definition.OperandWidths[i] {
		case 1:
			result[offset] = uint8(operand)
		case 2:
			binary.BigEndian.PutUint16(result[offset:], uint16(operand))
			break
		}

		offset += definition.OperandWidths[i]
	}

	return result
}
