package vm

import (
	"fmt"
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1, 1},
		{"2", 2, 1},
		{"1 + 2", 3, 1},
	}

	runVmTests(t, tests)
}

type vmTestCase struct {
	input          string
	expected       interface{}
	finalStackSize int
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		compiler := compiler.New()

		compiler.Compile(program)

		vm := New(compiler.Bytecode())

		err := vm.Execute()
		if err != nil {
			t.Fatalf("Failed to execute: %s", err)
		}

		if vm.stackPointer != test.finalStackSize {
			t.Fatalf("Final stack size %d not as expected (%d)", vm.stackPointer, test.finalStackSize)
		}

		stackItem := vm.StackTop()

		testExpectedObject(t, test.expected, stackItem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Fatalf("Test failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	converted, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("Object %v not integer but %T", actual, actual)
	}

	if converted.Value != expected {
		return fmt.Errorf(
			"Object value %d is wrong, expected %d",
			converted.Value, expected,
		)
	}

	return nil
}
