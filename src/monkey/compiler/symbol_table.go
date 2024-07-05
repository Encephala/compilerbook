package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
	BuiltinScope
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Parent *SymbolTable

	store            map[string]Symbol
	nonBuiltinsCount int
}

// Number of symbols defined (ignoring builtin functions)
func (st *SymbolTable) Len() int {
	return st.nonBuiltinsCount
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{nil, make(map[string]Symbol), 0}
}

func NewEnclosedSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{parent, make(map[string]Symbol), 0}
}

func (st *SymbolTable) Define(name string) Symbol {
	var scope SymbolScope

	if st.Parent == nil {
		scope = GlobalScope
	} else {
		scope = LocalScope
	}

	result := Symbol{
		Name:  name,
		Scope: scope,
		Index: st.Len(),
	}

	st.store[name] = result
	st.nonBuiltinsCount++

	return result
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Scope: BuiltinScope,
		Index: index,
	}

	st.store[name] = symbol

	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	result, ok := st.store[name]

	if !ok && st.Parent != nil {
		return st.Parent.Resolve(name)
	}

	return result, ok
}
