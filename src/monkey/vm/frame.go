package vm

import (
	"monkey/object"
	"monkey/opcode"
)

type Frame struct {
	function           *object.CompiledFunction
	instructionPointer int
	basePointer        int
}

func NewFrame(function *object.CompiledFunction, returnPointer int) *Frame {
	return &Frame{
		function:           function,
		instructionPointer: 0,
		basePointer:        returnPointer,
	}
}

func (frame *Frame) Instructions() *opcode.Instructions {
	return &frame.function.Instructions
}
