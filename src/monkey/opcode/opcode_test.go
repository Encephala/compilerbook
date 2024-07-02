package opcode

import "testing"

func TestMakeInstruction(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []int
		expected []byte
	}{
		{OpGetConstant, []int{65534}, []byte{byte(OpGetConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
	}

	for _, test := range tests {
		instruction := MakeInstruction(test.op, test.operands...)

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
		MakeInstruction(OpAdd),
		MakeInstruction(OpGetLocal, 1),
		MakeInstruction(OpGetConstant, 2),
		MakeInstruction(OpGetConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpGetLocal 1
0003 OpGetConstant 2
0006 OpGetConstant 65535
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
		{OpGetConstant, []int{65535}, 2},
		{OpGetLocal, []int{128}, 1},
	}

	for _, test := range tests {
		instruction := MakeInstruction(test.code, test.operands...)

		definition := Lookup(test.code)

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
