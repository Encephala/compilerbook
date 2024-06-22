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

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []opcode.Instruction
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		compiler := New()
		compiler.Compile(program)

		bytecode := compiler.Bytecode()

		concatenatedInstructions := concatInstructions(test.expectedInstructions)

		err := testInstructions(concatenatedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s\n", err)
		}

		err = testConstants(test.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s\n", err)
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

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
			},
		},
	}

	runCompilerTests(t, tests)
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
				"wrong instruction %d at %d, expected %d",
				actual[i], i, instruction,
			)
		}
	}

	return nil
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
