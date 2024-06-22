package compiler

import (
	"monkey/ast"
	"monkey/object"
	"monkey/opcode"
)

type Compiler struct {
	instructions opcode.Instructions
	constants    []object.Object
}

type Bytecode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: []byte{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
