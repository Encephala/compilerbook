package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {
			Name:  "a",
			Scope: GlobalScope,
			Index: 0,
		},
		"b": {
			Name:  "b",
			Scope: GlobalScope,
			Index: 1,
		},
	}

	table := NewSymbolTable()

	a := table.Define("a")
	if a != expected["a"] {
		t.Errorf("Wrong symbol %+v, expected %+v", expected["a"], a)
	}

	b := table.Define("b")
	if b != expected["b"] {
		t.Errorf("Wrong symbol %+v, expected %+v", expected["b"], b)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}
	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}
}
