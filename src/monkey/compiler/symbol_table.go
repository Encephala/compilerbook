package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
	BuiltinScope
	FreeScope
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

	FreeSymbols []Symbol
}

// Number of symbols defined (ignoring builtin functions)
func (st *SymbolTable) Len() int {
	return st.nonBuiltinsCount
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{nil, make(map[string]Symbol), 0, []Symbol{}}
}

func NewEnclosedSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{parent, make(map[string]Symbol), 0, []Symbol{}}
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
		result, ok = st.Parent.Resolve(name)
		if !ok {
			return result, false
		}

		if result.Scope == GlobalScope || result.Scope == BuiltinScope {
			return result, true
		}

		// Then, either local to parent symbols, or free in parent hence free here as well.
		free := st.defineFree(result)
		return free, true
	}

	return result, ok
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	symbol := Symbol{
		Name:  original.Name,
		Scope: FreeScope,
		Index: len(st.FreeSymbols) - 1,
	}

	st.store[symbol.Name] = symbol

	return symbol
}
