package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/object"
	"monkey/opcode"
	"monkey/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []opcode.Instruction
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 - 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpSubtract),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 * 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpMultiply),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 / 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpDivide),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "-69",
			expectedConstants: []interface{}{69},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpNegate),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true; false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpPushTrue),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpPushFalse),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 == 1",
			expectedConstants: []interface{}{1, 1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpEquals),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "2 > 1",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpGreaterThan),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "2 < 1",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpGreaterThan),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "!true; true; !false;",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpPushTrue),
				opcode.MakeInstruction(opcode.OpLogicalNot),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpPushTrue),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpPushFalse),
				opcode.MakeInstruction(opcode.OpLogicalNot),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) { 69 }",
			expectedConstants: []interface{}{69},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpPushTrue),
				opcode.MakeInstruction(opcode.OpJumpNotTruthy, 10),
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpJump, 11),
				opcode.MakeInstruction(opcode.OpPushNull),
				// Alternative implicitly returns null, so still should pop the ExpressionStatement result
				// But no alternative means no jump needed at end of consequence
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "if (false) { 69 } else { 420 }",
			expectedConstants: []interface{}{69, 420},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpPushFalse),
				opcode.MakeInstruction(opcode.OpJumpNotTruthy, 10),
				opcode.MakeInstruction(opcode.OpReadConstant, 0),
				opcode.MakeInstruction(opcode.OpJump, 13),
				opcode.MakeInstruction(opcode.OpReadConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	for _, test := range tests {
		program := parse(test.input)

		compiler := New()
		compiler.Compile(program)

		bytecode := compiler.Bytecode()

		concatenatedInstructions := concatInstructions(test.expectedInstructions)

		err := testConstants(test.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s\n", err)
		}

		err = testInstructions(concatenatedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s\n", err)
		}
	}
}

func concatInstructions(instructions []opcode.Instruction) opcode.Instructions {
	var result opcode.Instructions

	for _, instruction := range instructions {
		result = append(result, instruction...)
	}

	return result
}

func testConstants(expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf(
			"wrong number of constants %d, expected %d",
			len(actual), len(expected),
		)
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf(
					"constant %d not correct: %s",
					i, err,
				)
			}
		}
	}

	return nil
}

func testInstructions(expected opcode.Instructions, actual opcode.Instructions) error {
	if len(expected) != len(actual) {
		return fmt.Errorf(
			"Wrong instructions %q, expected %q", actual, expected,
		)
	}

	for i, instruction := range expected {
		if instruction != actual[i] {
			return fmt.Errorf(
				"wrong instruction at %d:\nactual: %q\nexpected: %q",
				i, actual, expected,
			)
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	converted, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("object %v not integer but %T", actual, actual)
	}

	if converted.Value != expected {
		return fmt.Errorf(
			"object value %d is wrong, expected %d",
			converted.Value, expected,
		)
	}

	return nil
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
