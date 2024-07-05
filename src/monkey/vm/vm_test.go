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
		{`"deez"`, "deez"},
		{`"deez" + " " + "nuts"`, "deez nuts"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[69]; [1 + 2 - 3]", []int{0}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"{}", map[int]int{}},
		{
			"{1: 2, 2: 3}",
			map[int]int{
				1: 2,
				2: 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[int]int{
				2: 4,
				6: 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][1 + 1]", 3},
		{"[[1, 2, 3]][0][1]", 2},
		{"[][0]", Null},
		{"[1, 2, 3][4]", Null},
		{"[1, 2, 3][-1]", Null},
		{"{1: 2, 3: 4}[1]", 2},
		{"{}[0]", Null},
		{"{1: 2, 3: 4}[0]", Null},
	}

	runVmTests(t, tests)
}

func TestFunctionCallsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `let function = fn() { 5 + 10; };
			function()`,
			expected: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two()
			`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { 1 };
			let b = fn() { a() + 1 };
			let c = fn() { b() + 1 };
			c();
			`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { return 1; 2 };
			a();
			`,
			expected: 1,
		},
		{
			input: `
			let a = fn() { if (false) { return 69 } else { return 420 } };
			a()
			`,
			expected: 420,
		},
		{
			input: `
			let a = fn() { };
			a()
			`,
			expected: Null,
		},
		{
			input: `
			let noReturn = fn() { };
			let noReturnTwo = fn() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			expected: Null,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `let identity = fn(a) { a; };
			identity(4)`,
			expected: 4,
		},
		{
			input: `let sum = fn(a, b) { a + b; };
			sum(1, 2);`,
			expected: 3,
		},
		{
			input: `let sum = fn(a, b) {
				let c = a + b;
				return c;
			};
			sum(1, 2)`,
			expected: 3,
		},
		{
			input: `let c = 10;
			let sum = fn(a, b) { return a + b + c; }
			sum(1, 2) + sum(3, 4);`,
			expected: 30,
		},
		{
			input: `let sum = fn(a, b) { return a + b; };
			let outer = fn() { return sum(1, 2) + sum(3, 4); };
			outer();`,
			expected: 10,
		},
		{
			input: `let one = fn() { 1; };
			let sum = fn(a, b) { return a + b; };
			sum(one() + one(), one())`,
			expected: 3,
		},
		{
			input: `let one = fn() { 1; };
			let sum = fn(a, b) { return a + b; };
			sum(sum(one(), one()), one())`,
			expected: 3,
		},
		{
			input: `let global = 10;

			let sum = fn(a, b) {
				let c = a + b;
				return c + global;
			};

			let outer = fn() {
				return sum(1, 2) + sum(3, 4) + global;
			};

			outer() + global;`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "fn() { 1 }(2)",
			expected: "wrong number of arguments 1, expected 0",
		},
		{
			input:    "fn(a) { a }()",
			expected: "wrong number of arguments 0, expected 1",
		},
		{
			input:    "fn(a, b) { a; b }(1)",
			expected: "wrong number of arguments 1, expected 2",
		},
	}

	for _, test := range tests {
		program := parse(test.input)

		c := compiler.New()
		err := c.Compile(program)
		if err != nil {
			t.Fatalf("compiler error :%s", err)
		}

		vm := New(c.Bytecode())
		err = vm.Execute()
		if err == nil {
			t.Fatalf("expected error but didn't get one")
		}

		if err.Error() != test.expected {
			t.Fatalf("error %q is wrong, expected %q", err.Error(), test.expected)
		}
	}
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let oneReturnerReturner = fn() { fn() { 1 }; };
			oneReturnerReturner()();
			`,
			expected: 1,
		},
		{
			input: `
			let oneReturner = fn() { 1; };
			let oneReturnerReturner = fn() { oneReturner; };
			oneReturnerReturner()();
			`,
			expected: 1,
		},
		{
			input: `
			let oneReturnerReturner = fn() {
				let oneReturner = fn() { 1 };
				oneReturner;
			}
			oneReturnerReturner()();
			`,
			expected: 1,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `let one = fn() { let one = 1; one };
					one();`,
			expected: 1,
		},
		{
			input: `let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			oneAndTwo()`,
			expected: 3,
		},
		{
			input: `
				let first = fn() { let a = 50; a; };
				let second = fn() { let a = 100; a; };
				first() + second();
				`,
			expected: 150,
		},
		{
			input: `
				let globalSeed = 50;
				let minusOne = fn() {
				let num = 1;
				globalSeed - num;
				}
				let minusTwo = fn() {
				let num = 2;
				globalSeed - num;
				}
				minusOne() + minusTwo();
			`,
			expected: 97,
		},
	}

	runVmTests(t, tests)
}

func TestBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{`len("one", "two")`,
			&object.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`puts("hello", "world!")`, Null}, // Lmayo this actually prints to console if tests fail
		{`first([1, 2, 3])`, 1},
		{`first([])`, Null},
		{`first(1)`,
			&object.Error{
				Message: "argument to `first` must be ARRAY, got INTEGER",
			},
		},
		{`last([1, 2, 3])`, 3},
		{`last([])`, Null},
		{`last(1)`,
			&object.Error{
				Message: "argument to `last` must be ARRAY, got INTEGER",
			},
		},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, Null},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be ARRAY, got INTEGER",
			},
		},
	}

	runVmTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `let newClosure = fn(a) { fn() { a; } };
			let closure = newClosure(69);
			closure();`,
			expected: 69,
		},
		{
			input: `let newAdder = fn(a, b) {
				fn(c) { a + b + c };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `let newAdder = fn(a, b) {
				let c = a + b;
				fn(d) { c + d };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `let newAdderOuter = fn(a, b) {
				let c = a + b;
				fn(d) {
					let e = d + c;
					fn(f) { e + f; };
				};
			};

			let newAdderInner = newAdderOuter(1, 2)
			let adder = newAdderInner(3);
			adder(8);
			`,
			expected: 14,
		},
		{
			input: `let a = 1;
			let newAdderOuter = fn(b) {
				fn(c) {
					fn(d) { a + b + c + d };
				};
			};

			let newAdderInner = newAdderOuter(2)
			let adder = newAdderInner(3);
			adder(8);
			`,
			expected: 14,
		},
		{
			input: `let newClosure = fn(a, b) {
				let one = fn() { a; };
				let two = fn() { b; };
				fn() { one() + two(); };
			};
			let closure = newClosure(9, 90);
			closure();
			`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `let countDown = fn(start) {
				if (start == 0) { return 0; }
				else { countDown(start - 1); }
			};

			countDown(1);`,
			expected: 0,
		},
		{
			input: `let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};

			let wrapper = fn() {
				countDown(1);
			};
			wrapper();
			`,
			expected: 0,
		},
		{
			input: `let wrapper = fn() {
				let countDown = fn(x) {
					if (x == 0) {
					return 0;
					} else {
					countDown(x - 1);
					}
				};

				countDown(1);
			};

			wrapper();
			`,
			expected: 0,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fibonacci = fn(x) {
				if (x == 0) { return 0; }

				if (x == 1) { return 1; }

				return fibonacci(x - 1) + fibonacci(x - 2);
			};

			fibonacci(15);
			`,
			expected: 610,
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
			t.Fatalf("Failed to compile: %s\n", err)
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

	case map[int]int:
		hash, ok := actual.(*object.Hash)

		if !ok {
			t.Fatalf("Object %v not hash but %T", actual, actual)
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("Wrong number of elements %d, expected %d", len(hash.Pairs), len(expected))
		}

		for key, value := range expected {
			pair, ok := hash.Pairs[(&object.Integer{Value: int64(key)}).HashKey()]

			if !ok {
				t.Errorf("Key %v not found in hash", key)
				continue
			}

			err := testIntegerObject(int64(value), pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s\n", err)
			}
		}

	case *object.Error:
		converted, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("Object %+v is not an error but %T", actual, actual)
		}

		if converted.Message != expected.Message {
			t.Errorf("Wrong error message %q, expected %q", converted.Message, expected.Message)
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
