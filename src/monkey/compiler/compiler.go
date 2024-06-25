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

	lastInstruction     *EmittedInstruction
	previousInstruction *EmittedInstruction // So we can set lastInstruction after popping off an instruction
}

type EmittedInstruction struct {
	code  opcode.OpCode
	index int
}

type Bytecode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: []byte{},
		constants:    []object.Object{},

		lastInstruction:     nil,
		previousInstruction: nil,
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
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

	case *ast.IfExpression:
		c.Compile(node.Condition)
		c.emit(opcode.OpJumpNotTruthy, -1) // Invalid jump location as temporary value

		indexJumpNotTruthy := c.lastInstruction.index

		c.Compile(node.Consequence)
		// What's null safety? I hardly know her
		if c.lastInstruction.code == opcode.OpPop {
			c.removeLastPop()
		}

		c.emit(opcode.OpJump, -1) // Invalid jump location as temporary value

		c.replaceInstruction(indexJumpNotTruthy, opcode.MakeInstruction(opcode.OpJumpNotTruthy, len(c.instructions)))

		indexJump := c.lastInstruction.index

		c.Compile(node.Alternative)
		if c.lastInstruction.code == opcode.OpPop {
			c.removeLastPop()
		}

		c.replaceInstruction(indexJump, opcode.MakeInstruction(opcode.OpJump, len(c.instructions)))

	case *ast.BlockStatement:
		if node == nil {
			c.emit(opcode.OpPushNull)

			return
		}

		for _, statement := range node.Statements {
			c.Compile(statement)
		}

	case *ast.InfixExpression:
		if node.Operator[0] == byte('<') {
			// Switch order of operands, so we can reuse OpGreater
			c.Compile(node.Right)

			c.Compile(node.Left)
		} else {
			c.Compile(node.Left)

			c.Compile(node.Right)
		}

		switch node.Operator {
		case "+":
			c.emit(opcode.OpAdd)
		case "-":
			c.emit(opcode.OpSubtract)
		case "*":
			c.emit(opcode.OpMultiply)
		case "/":
			c.emit(opcode.OpDivide)

		case "==":
			c.emit(opcode.OpEquals)
		case "!=":
			c.emit(opcode.OpNotEquals)
		case ">":
			c.emit(opcode.OpGreaterThan)
		case "<":
			c.emit(opcode.OpGreaterThan)

		default:
			panic(fmt.Sprintf("Invalid infix operator: %q", node.Operator))
		}

	case *ast.PrefixExpression:
		c.Compile(node.Right)
		switch node.Operator {
		case "-":
			c.emit(opcode.OpNegate)

		case "!":
			c.emit(opcode.OpLogicalNot)

		default:
			panic(fmt.Sprintf("Invalid prefix operator: %q", node.Operator))
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}

		index := c.addConstant(integer)

		c.emit(opcode.OpReadConstant, index)

	case *ast.Boolean:
		if node.Value {
			c.emit(opcode.OpPushTrue)
		} else {
			c.emit(opcode.OpPushFalse)
		}

	default:
		panic(fmt.Sprintf("Invalid node type: %T", node))
	}
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) {
	bytecode := opcode.MakeInstruction(op, operands...)

	starting_position := len(c.instructions)
	c.instructions = append(c.instructions, bytecode...)

	c.lastInstruction = &EmittedInstruction{
		code:  op,
		index: starting_position,
	}
}

func (c *Compiler) addConstant(constant object.Object) int {
	constantIndex := len(c.constants)

	c.constants = append(c.constants, constant)

	return constantIndex
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:len(c.instructions)-1]

	c.lastInstruction = c.previousInstruction
	c.previousInstruction = nil
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[position+i] = newInstruction[i]
	}
}
