package opcode

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []uint
		expected []byte
	}{
		{OpConstant, []uint{65534}, []byte{byte(OpConstant), 255, 254}},
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
