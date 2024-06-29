package vm

import (
	"monkey/object"
	"monkey/opcode"
)

type Frame struct {
	function           *object.CompiledFunction
	instructionPointer int
}

func NewFrame(function *object.CompiledFunction) *Frame {
	return &Frame{
		function:           function,
		instructionPointer: 0,
	}
}

func (frame *Frame) Instructions() *opcode.Instructions {
	return &frame.function.Instructions
}
