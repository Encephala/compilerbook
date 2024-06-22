package opcode

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		if len(instruction) != len(test.expected) {
			t.Errorf(
				"instruction has wrong length %d but expected %d",
				len(instruction), len(test.expected),
			)
		}

		for i, b := range test.expected {
			if instruction[i] != b {
				t.Errorf(
					"Wrong byte at %v: %d but expected %d",
					i, instruction[i], b,
				)
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instruction{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatenated := Instructions{}
	for _, instruction := range instructions {
		concatenated = append(concatenated, instruction...)
	}

	if concatenated.String() != expected {
		t.Fatalf(
			"Wrong string %q, expected %q\n",
			concatenated.String(), expected,
		)
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		code      OpCode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, test := range tests {
		instruction := Make(test.code, test.operands...)

		definition, err := Lookup(test.code)
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		operandsRead, bytesRead := ReadOperands(definition, instruction[1:])
		if bytesRead != test.bytesRead {
			t.Fatalf(
				"Number of bytes %d wrong, expected %d",
				bytesRead, test.bytesRead,
			)
		}

		for i, expected := range operandsRead {
			if operandsRead[i] != expected {
				t.Errorf(
					"wrong operand %d, expected %d",
					operandsRead[i], expected,
				)
			}
		}
	}
}
