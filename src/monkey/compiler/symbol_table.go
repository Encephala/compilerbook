package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	parent *SymbolTable
	store  map[string]Symbol
}

func (st *SymbolTable) Len() int {
	return len(st.store)
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{nil, make(map[string]Symbol)}
}

func NewEnclosedSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{parent, make(map[string]Symbol)}
}

func (st *SymbolTable) Define(name string) Symbol {
	var scope SymbolScope

	if st.parent == nil {
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

	return result
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	result, ok := st.store[name]

	if !ok && st.parent != nil {
		return st.parent.Resolve(name)
	}

	return result, ok
}
