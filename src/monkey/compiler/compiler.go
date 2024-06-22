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
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}

		c.addConstant(integer)
	}

	return nil
}

func (c *Compiler) addConstant(constant object.Object) {
	index := len(c.constants)

	bytecode := opcode.Make(opcode.OpConstant, uint(index))

	c.constants = append(c.constants, constant)

	for _, b := range bytecode {
		c.instructions = append(c.instructions, b)
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
