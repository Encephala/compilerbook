package vm

import (
	"monkey/object"
	"monkey/opcode"
)

type Frame struct {
	closure            *object.Closure
	instructionPointer int
	basePointer        int
}

func NewFrame(closure *object.Closure, returnPointer int) *Frame {
	return &Frame{
		closure:            closure,
		instructionPointer: 0,
		basePointer:        returnPointer,
	}
}

func (frame *Frame) Instructions() *opcode.Instructions {
	return &frame.closure.Function.Instructions
}
