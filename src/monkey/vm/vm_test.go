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

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"1 / 2", 0},
		{"6 / 2", 3},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},

		{"1 == 1", true},
		{"2 == 1", false},
		{"2 != 1", true},
		{"2 != 2", false},
		{"true == true == true", true},
		{"true == false", false},
		{"true == false != false", false},
		{"true != false", true},
		{"true != true", false},

		{"2 > 1", true},
		{"2 > 1 == false", false},
		{"2 > 2", false},
		{"2 > 2 == false", true},

		{"1 < 2", true},
		{"2 < 2", false},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	for _, test := range tests {
		program := parse(test.input)

		compiler := compiler.New()

		compiler.Compile(program)

		vm := New(compiler.Bytecode())

		err := vm.Execute()
		if err != nil {
			t.Fatalf("Failed to execute: %s", err)
		}

		result := vm.LastStackTop()

		testExpectedObject(t, test.expected, result)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Fatalf("Test failed: %s", err)
		}

	case bool:
		err := testBoolObject(bool(expected), actual)
		if err != nil {
			t.Fatalf("Test failed: %s", err)
		}

	default:
		panic(fmt.Sprintf("Unimplemented: %T", expected))
	}
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

func testBoolObject(expected bool, actual object.Object) error {
	converted, ok := actual.(*object.Boolean)

	if !ok {
		return fmt.Errorf("Object %v not boolean but %T", actual, actual)
	}

	if converted.Value != expected {
		return fmt.Errorf(
			"Object value %t is wrong, expected %t",
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
