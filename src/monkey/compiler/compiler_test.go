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
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 - 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpSubtract),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 * 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpMultiply),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "1 / 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpDivide),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "-69",
			expectedConstants: []interface{}{69},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
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
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpEquals),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "2 > 1",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGreaterThan),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "2 < 1",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
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
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
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
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpJump, 13),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let one = 1;
			let two = 2;`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpSetGlobal, 1),
			},
		},
		{
			input: `
			let one = 1;
			one;`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetGlobal, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `
			let one = 1;
			let two = one;
			two;`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetGlobal, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 1),
				opcode.MakeInstruction(opcode.OpGetGlobal, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"deez"`,
			expectedConstants: []interface{}{"deez"},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             `"deez" + "nuts"`,
			expectedConstants: []interface{}{"deez", "nuts"},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiteral(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpArray, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpArray, 3),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "[1 + 2 - 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpSubtract),
				opcode.MakeInstruction(opcode.OpArray, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpHash, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2, 5: 6, 3: 4}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpGetConstant, 4),
				opcode.MakeInstruction(opcode.OpGetConstant, 5),
				opcode.MakeInstruction(opcode.OpHash, 3),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpGetConstant, 4),
				opcode.MakeInstruction(opcode.OpGetConstant, 5),
				opcode.MakeInstruction(opcode.OpMultiply),
				opcode.MakeInstruction(opcode.OpHash, 2),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpArray, 3),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpGetConstant, 4),
				opcode.MakeInstruction(opcode.OpAdd),
				opcode.MakeInstruction(opcode.OpIndex),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpHash, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpSubtract),
				opcode.MakeInstruction(opcode.OpIndex),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { return 5 + 10 };",
			expectedConstants: []interface{}{
				5,
				10,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 0),
					opcode.MakeInstruction(opcode.OpGetConstant, 1),
					opcode.MakeInstruction(opcode.OpAdd),
					opcode.MakeInstruction(opcode.OpReturnValue),
				}},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: "fn() { 1; 2 };",
			expectedConstants: []interface{}{
				1,
				2,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 0),
					opcode.MakeInstruction(opcode.OpPop),
					opcode.MakeInstruction(opcode.OpGetConstant, 1),
					opcode.MakeInstruction(opcode.OpReturnValue),
				}},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: "fn() {  };",
			expectedConstants: []interface{}{
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpReturn),
				}},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex %d wrong, expected 0", compiler.scopeIndex)
	}
	globalSymbols := compiler.symbols

	compiler.emit(opcode.OpMultiply)

	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex %d wrong, expected 1", compiler.scopeIndex)
	}

	compiler.emit(opcode.OpSubtract)

	currentInstructions := compiler.currentInstructions()
	if len(*currentInstructions) != 1 {
		t.Errorf("instructions length %d wrong, expected 1", len(*currentInstructions))
	}

	lastInstruction := compiler.currentScope().lastInstruction
	if lastInstruction.code != opcode.OpSubtract {
		t.Errorf("opcode %d wrong, expected %d", lastInstruction.code, opcode.OpSubtract)
	}

	if compiler.symbols.Parent != globalSymbols {
		t.Errorf("compiler's local symbol table has incorrect parent")
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex %d wrong, expected 0", compiler.scopeIndex)
	}

	if compiler.symbols != globalSymbols || compiler.symbols == nil || compiler.symbols.Parent != nil {
		t.Errorf("compiler did not set symbol table on scope leave")
	}

	compiler.emit(opcode.OpAdd)

	currentInstructions = compiler.currentInstructions()
	if len(*currentInstructions) != 2 {
		t.Errorf("instructions length %d wrong, expected 2", len(*currentInstructions))
	}

	lastInstruction = compiler.currentScope().lastInstruction
	if lastInstruction.code != opcode.OpAdd {
		t.Errorf("last opcode %d wrong, expected %d", lastInstruction.code, opcode.OpAdd)
	}

	previousInstruction := compiler.currentScope().previousInstruction
	if previousInstruction.code != opcode.OpMultiply {
		t.Errorf("previous opcode %d wrong, expected %d", lastInstruction.code, opcode.OpMultiply)
	}
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { 24 }();",
			expectedConstants: []interface{}{
				24,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 0),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpCall, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `let function = fn() { 24 };
			function()`,
			expectedConstants: []interface{}{
				24,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 0),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetGlobal, 0),
				opcode.MakeInstruction(opcode.OpCall, 0),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `let function = fn(a) { a };
			function(69);`,
			expectedConstants: []interface{}{
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetLocal, 0),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
				69,
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpCall, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `let function = fn(a, b, c) { a; b; c };
			function(69, 420, 0);`,
			expectedConstants: []interface{}{
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetLocal, 0),
					opcode.MakeInstruction(opcode.OpPop),
					opcode.MakeInstruction(opcode.OpGetLocal, 1),
					opcode.MakeInstruction(opcode.OpPop),
					opcode.MakeInstruction(opcode.OpGetLocal, 2),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
				69,
				420,
				0,
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpGetConstant, 2),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpCall, 3),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLetStatementsScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `let num = 55;
					fn() { num }`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetGlobal, 0),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `fn() {
			let num = 55;
			num
			}`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 0),
					opcode.MakeInstruction(opcode.OpSetLocal, 0),
					opcode.MakeInstruction(opcode.OpGetLocal, 0),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 1),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `let c = 99;
			fn() {
			let a = 55;
			let b = 77;
			a + b + c
			}`,
			expectedConstants: []interface{}{
				99,
				55,
				77,
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetConstant, 1),
					opcode.MakeInstruction(opcode.OpSetLocal, 0),
					opcode.MakeInstruction(opcode.OpGetConstant, 2),
					opcode.MakeInstruction(opcode.OpSetLocal, 1),
					opcode.MakeInstruction(opcode.OpGetLocal, 0),
					opcode.MakeInstruction(opcode.OpGetLocal, 1),
					opcode.MakeInstruction(opcode.OpAdd),
					opcode.MakeInstruction(opcode.OpGetGlobal, 0),
					opcode.MakeInstruction(opcode.OpAdd),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
				opcode.MakeInstruction(opcode.OpSetGlobal, 0),
				opcode.MakeInstruction(opcode.OpGetConstant, 3),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `len([]);
			push([], 1)`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetBuiltin, 0),
				opcode.MakeInstruction(opcode.OpArray),
				opcode.MakeInstruction(opcode.OpCall, 1),
				opcode.MakeInstruction(opcode.OpPop),
				opcode.MakeInstruction(opcode.OpGetBuiltin, 5),
				opcode.MakeInstruction(opcode.OpArray),
				opcode.MakeInstruction(opcode.OpGetConstant),
				opcode.MakeInstruction(opcode.OpCall, 2),
				opcode.MakeInstruction(opcode.OpPop),
			},
		},
		{
			input: `fn() { len([]) }`,
			expectedConstants: []interface{}{
				[]opcode.Instruction{
					opcode.MakeInstruction(opcode.OpGetBuiltin, 0),
					opcode.MakeInstruction(opcode.OpArray, 0),
					opcode.MakeInstruction(opcode.OpCall, 1),
					opcode.MakeInstruction(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instruction{
				opcode.MakeInstruction(opcode.OpGetConstant, 0),
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
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("Compilation failed: %s\n", err)
		}

		bytecode := compiler.Bytecode()

		concatenatedInstructions := concatInstructions(test.expectedInstructions)

		err = testConstants(test.expectedConstants, bytecode.Constants)
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

func testInstructions(expected opcode.Instructions, actual opcode.Instructions) error {
	if len(expected) != len(actual) {
		return fmt.Errorf(
			"Wrong instructions:\n%q,\n\nexpected:\n%q", actual, expected,
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

		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf(
					"constant %d not correct: %s",
					i, err,
				)
			}

		case []opcode.Instruction:
			var concatenatedInstructions opcode.Instructions

			for _, instruction := range constant {
				concatenatedInstructions = append(concatenatedInstructions, instruction...)
			}

			convertedActual, ok := actual[i].(*object.CompiledFunction)

			if !ok {
				return fmt.Errorf(
					"constant %d not a function, got %T",
					i, actual[i],
				)
			}

			err := testInstructions(concatenatedInstructions, convertedActual.Instructions)
			if err != nil {
				return fmt.Errorf(
					"constant %d not correct: %s",
					i, err,
				)
			}

		default:
			panic(fmt.Sprintf("Invalid test type %T", constant))
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

func testStringObject(expected string, actual object.Object) error {
	converted, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("object %v not string but %T", actual, actual)
	}

	if converted.Value != expected {
		return fmt.Errorf(
			"object value %q is wrong, expected %q",
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
