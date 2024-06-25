package compiler

import (
	"fmt"
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

func (c *Compiler) Compile(node ast.Node) {
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			c.Compile(statement)
		}

	case *ast.ExpressionStatement:
		c.Compile(node.Expression)

		c.emit(opcode.OpPop)

	case *ast.InfixExpression:
		c.Compile(node.Left)

		c.Compile(node.Right)

		switch node.Operator {
		case "+":
			c.emit(opcode.OpAdd)
		case "-":
			c.emit(opcode.OpSubtract)
		case "*":
			c.emit(opcode.OpMultiply)
		case "/":
			c.emit(opcode.OpDivide)

		default:
			panic(fmt.Sprintf("Invalid infix operator: %q", node.Operator))
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}

		c.addConstant(integer)

	default:
		panic(fmt.Sprintf("Invalid node type: %T", node))
	}
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) {
	bytecode := opcode.Make(op, operands...)

	for _, b := range bytecode {
		c.instructions = append(c.instructions, b)
	}
}

func (c *Compiler) addConstant(constant object.Object) int {
	constantIndex := len(c.constants)

	c.constants = append(c.constants, constant)

	instructionIndex := len(c.instructions)

	c.emit(opcode.OpConstant, constantIndex)

	return instructionIndex
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
