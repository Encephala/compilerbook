package compiler

import (
	"testing"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("Wrong symbol %+v, expected %+v", a, expected["a"])
	}

	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("Wrong symbol %+v, expected %+v", b, expected["b"])
	}

	firstLocal := NewEnclosedSymbolTable(global)
	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("Wrong symbol %+v, expected %+v", c, expected["c"])
	}
	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("Wrong symbol %+v, expected %+v", d, expected["d"])
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("Wrong symbol %+v, expected %+v", e, expected["e"])
	}
	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("Wrong symbol %+v, expected %+v", f, expected["f"])
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

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, symbol := range expected {
		resolved, ok := local.Resolve(symbol.Name)
		if !ok {
			t.Errorf("Name %s wasn't found", symbol.Name)
		}

		if resolved != symbol {
			t.Errorf("Symbol %+v resolved wrong, expected %+v",
				resolved, symbol)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, test := range tests {
		for _, symbol := range test.expectedSymbols {
			resolved, ok := test.table.Resolve(symbol.Name)
			if !ok {
				t.Errorf("Name %s wasn't found", symbol.Name)
			}

			if resolved != symbol {
				t.Errorf("Symbol %+v resolved wrong, expected %+v",
					resolved, symbol)
			}
		}
	}
}

func TestBuiltin(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{
			Name:  "a",
			Scope: BuiltinScope,
			Index: 0,
		},
		{
			Name:  "b",
			Scope: BuiltinScope,
			Index: 1,
		},
		{
			Name:  "e",
			Scope: BuiltinScope,
			Index: 2,
		},
		{
			Name:  "f",
			Scope: BuiltinScope,
			Index: 3,
		},
	}

	for i, value := range expected {
		global.DefineBuiltin(i, value.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, symbol := range expected {
			result, ok := table.Resolve(symbol.Name)

			if !ok {
				t.Errorf("name %s not resolvable", symbol.Name)
			}

			if result != symbol {
				t.Errorf("Symbol %s resolved to %+v, expected %+v", symbol.Name, result, symbol)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}
	for _, test := range tests {
		for _, symbol := range test.expectedSymbols {
			result, ok := test.table.Resolve(symbol.Name)
			if !ok {
				t.Errorf("name %s not resolvable", symbol.Name)
				continue
			}
			if result != symbol {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					symbol.Name, symbol, result)
			}
		}

		if len(test.table.FreeSymbols) != len(test.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d",
				len(test.table.FreeSymbols), len(test.expectedFreeSymbols))
			continue
		}

		for i, symbol := range test.expectedFreeSymbols {
			result := test.table.FreeSymbols[i]
			if result != symbol {
				t.Errorf("wrong free symbol. got=%+v, want=%+v",
					result, symbol)
			}
		}
	}
}

func TestUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, symbol := range expected {
		resolved, ok := secondLocal.Resolve(symbol.Name)
		if !ok {
			t.Errorf("Couldn't resolve symbol %s", symbol.Name)
		}

		if resolved != symbol {
			t.Errorf("Symbol %+v wrong, expected %+v", resolved, symbol)
		}
	}

	expectedUnresolvable := []string{"b", "d"}

	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)

		if ok {
			t.Errorf("Resolved %s, this shouldn't happen", name)
		}
	}
}
