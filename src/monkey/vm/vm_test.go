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

		{"-69", -69},

		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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

		{"!true", false},
		{"(!true == false) == true", true},
		{"!5", false},
		{"!!true", true},
		{"!!5", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) {}", Null},
		{"if (true) { 69 }", 69},
		{"if (false) { 69 } else { 420 }", 420},
		{"if (false) { 69 }", Null},

		{"if (6 > 9) { 69 } else { 420 }", 420},
		{"if (10 > 9) { 69 } else { 420 }", 69},

		{"if (5) { 69 } else { 420 }", 69},
		{"if (0) { 69 } else { 420 }", 420},

		{"if (if (true) {}) { 69 } else { 420 }", 420},
		{"if (!if (true) {}) { 69 } else { 420 }", 69},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one;", 1},
		{"let one = 1; let two = 2; one + two;", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{
			`"deez"`,
			"deez",
		},
		{
			`"deez" + " " + "nuts"`,
			"deez nuts",
		},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"[]",
			[]int{},
		},
		{
			"[1, 2, 3]",
			[]int{1, 2, 3},
		},
		{
			"[69]; [1 + 2 - 3]",
			[]int{0},
		},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	for _, test := range tests {
		program := parse(test.input)

		compiler := compiler.New()

		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("Failed to execute: %s\n", err)
		}

		vm := New(compiler.Bytecode())

		err = vm.Execute()
		if err != nil {
			t.Fatalf("Failed to execute: %s\n", err)
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
		err := testBoolObject(expected, actual)
		if err != nil {
			t.Fatalf("Test failed: %s", err)
		}

	case *object.Null:
		if actual != Null {
			t.Fatalf("Object %v is not Null", actual)
		}

	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Fatalf("Test failed: %s", err)
		}

	case []int:
		array, ok := actual.(*object.Array)

		if !ok {
			t.Fatalf("Object %v not array but %T", actual, actual)
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("Wrong number of elements %d, expected %d", len(array.Elements), len(expected))
		}

		for i, integer := range expected {
			err := testIntegerObject(int64(integer), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s\n", err)
			}
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

func testStringObject(expected string, actual object.Object) error {
	converted, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("Object %v not boolean but %T", actual, actual)
	}

	if converted.Value != expected {
		return fmt.Errorf(
			"Object value %q is wrong, expected %q",
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
