package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/opcode"
	"sort"
)

type Compiler struct {
	constants []object.Object

	symbols *SymbolTable

	scopes     []*CompilationScope
	scopeIndex int
}

type CompilationScope struct {
	instructions *opcode.Instructions

	lastInstruction     *EmittedInstruction
	previousInstruction *EmittedInstruction // So we can set lastInstruction after popping off an instruction
}

type EmittedInstruction struct {
	code  opcode.OpCode
	index int
}

func (c *Compiler) currentScope() *CompilationScope {
	return c.scopes[c.scopeIndex]
}

func (c *Compiler) currentInstructions() *opcode.Instructions {
	return c.currentScope().instructions
}

func (c *Compiler) enterScope() {
	scope := &CompilationScope{
		instructions: &opcode.Instructions{},

		lastInstruction:     nil,
		previousInstruction: nil,
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbols = NewEnclosedSymbolTable(c.symbols)
}

func (c *Compiler) leaveScope() opcode.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbols = c.symbols.Parent

	return *instructions
}

type Bytecode struct {
	Instructions opcode.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	mainScope := &CompilationScope{
		instructions:        &opcode.Instructions{},
		lastInstruction:     nil,
		previousInstruction: nil,
	}

	return &Compiler{
		constants: []object.Object{},

		symbols: NewSymbolTable(),

		scopes:     []*CompilationScope{mainScope},
		scopeIndex: 0,
	}
}

func NewWithState(constants []object.Object, symbols *SymbolTable) *Compiler {
	mainScope := &CompilationScope{
		instructions:        &opcode.Instructions{},
		lastInstruction:     nil,
		previousInstruction: nil,
	}

	return &Compiler{
		constants: constants,

		symbols: symbols,

		scopes:     []*CompilationScope{mainScope},
		scopeIndex: 0,
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: *c.currentInstructions(),
		Constants:    c.constants,
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

		c.emit(opcode.OpPop)

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		c.emit(opcode.OpJumpNotTruthy, -1) // Invalid jump location as temporary value

		indexJumpNotTruthy := c.currentScope().lastInstruction.index

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		// What's null safety? I hardly know her
		if c.currentScope().lastInstruction.code == opcode.OpPop {
			c.removeLastInstruction()
		}

		c.emit(opcode.OpJump, -1) // Invalid jump location as temporary value

		c.replaceInstruction(indexJumpNotTruthy, opcode.MakeInstruction(
			opcode.OpJumpNotTruthy,
			len(*c.currentInstructions()),
		))

		indexJump := c.currentScope().lastInstruction.index

		if node.Alternative == nil {
			c.emit(opcode.OpPushNull)
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return nil
			}
			if c.currentScope().lastInstruction.code == opcode.OpPop {
				c.removeLastInstruction()
			}
		}

		c.replaceInstruction(
			indexJump,
			opcode.MakeInstruction(opcode.OpJump, len(*c.currentInstructions())),
		)

	case *ast.BlockStatement:
		if len(node.Statements) == 0 {
			c.emit(opcode.OpPushNull)

			return nil
		}

		var err error
		for _, statement := range node.Statements {
			err = c.Compile(statement)
			if err != nil {
				return nil
			}
		}

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return nil
		}

		symbol := c.symbols.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpSetGlobal, symbol.Index)
		} else {
			c.emit(opcode.OpSetLocal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbols.Resolve(node.Value)

		if !ok {
			return fmt.Errorf("Symbol %q not found", node.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpGetGlobal, symbol.Index)
		} else {
			c.emit(opcode.OpGetLocal, symbol.Index)
		}

	case *ast.InfixExpression:
		if node.Operator[0] == byte('<') {
			// Switch order of operands, so we can reuse OpGreater
			err := c.Compile(node.Right)
			if err != nil {
				return nil
			}

			err = c.Compile(node.Left)
			if err != nil {
				return nil
			}
		} else {
			err := c.Compile(node.Left)
			if err != nil {
				return nil
			}

			err = c.Compile(node.Right)
			if err != nil {
				return nil
			}
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
		err := c.Compile(node.Right)
		if err != nil {
			return nil
		}

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

		c.emit(opcode.OpGetConstant, index)

	case *ast.Boolean:
		if node.Value {
			c.emit(opcode.OpPushTrue)
		} else {
			c.emit(opcode.OpPushFalse)
		}

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}

		index := c.addConstant(str)

		c.emit(opcode.OpGetConstant, index)

	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}

		for key := range node.Pairs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			err := c.Compile(key)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[key])
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpHash, len(node.Pairs))

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(opcode.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Implicit return, replace last pop with a return
		if c.lastInstructionIs(opcode.OpPop) {
			c.replaceInstruction(
				len(*c.currentInstructions())-1,
				opcode.MakeInstruction(opcode.OpReturnValue),
			)
		}
		// Empty body
		if c.lastInstructionIs(opcode.OpPushNull) {
			c.replaceInstruction(
				len(*c.currentInstructions())-1,
				opcode.MakeInstruction(opcode.OpReturn),
			)
		}
		// Last statement is a let
		if c.lastInstructionIs(opcode.OpSetGlobal) {
			c.emit(opcode.OpReturn)
		}

		numberOfLocals := c.symbols.Len()

		instructions := c.leaveScope()
		result := &object.CompiledFunction{Instructions: instructions, NumberOfLocals: numberOfLocals}
		index := c.addConstant(result)

		c.emit(opcode.OpGetConstant, index)

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return nil
		}

		c.emit(opcode.OpReturnValue)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return nil
		}

		c.emit(opcode.OpCall)

	default:
		panic(fmt.Sprintf("Invalid node type: %T", node))
	}

	return nil
}

func (c *Compiler) emit(op opcode.OpCode, operands ...int) {
	bytecode := opcode.MakeInstruction(op, operands...)

	currentInstructions := c.currentInstructions()

	starting_position := len(*currentInstructions)
	*currentInstructions = append(*currentInstructions, bytecode...)

	c.currentScope().previousInstruction = c.currentScope().lastInstruction
	c.currentScope().lastInstruction = &EmittedInstruction{
		code:  op,
		index: starting_position,
	}
}

func (c *Compiler) addConstant(constant object.Object) int {
	constantIndex := len(c.constants)

	c.constants = append(c.constants, constant)

	return constantIndex
}

func (c *Compiler) lastInstructionIs(operation opcode.OpCode) bool {
	return c.currentScope().lastInstruction.code == operation
}

func (c *Compiler) removeLastInstruction() {
	currentInstructions := c.currentInstructions()

	*currentInstructions = (*currentInstructions)[:len(*currentInstructions)-1]

	c.currentScope().lastInstruction = c.scopes[c.scopeIndex].previousInstruction
	c.currentScope().previousInstruction = nil
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		(*c.currentScope().instructions)[position+i] = newInstruction[i]
	}
}
