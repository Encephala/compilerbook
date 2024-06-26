package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = 0
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store map[string]Symbol
}

func (st *SymbolTable) Len() int {
	return len(st.store)
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{make(map[string]Symbol)}
}

func (st *SymbolTable) Define(name string) Symbol {
	// TODO: insert in non-global scopes

	result := Symbol{
		Name:  name,
		Scope: GlobalScope,
		Index: st.Len(),
	}

	st.store[name] = result

	return result
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	result, ok := st.store[name]

	return result, ok
}
