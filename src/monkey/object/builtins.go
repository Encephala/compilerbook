package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		Name: "len",
		Builtin: &Builtin{
			func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{
						fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)),
					}
				}

				switch arg := args[0].(type) {
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}

				case *String:
					return &Integer{Value: int64(len(arg.Value))}

				default:
					return &Error{
						fmt.Sprintf("argument to `len` not supported, got %s", arg.Type()),
					}
				}
			},
		},
	},
	{
		Name: "puts",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}

				return nil
			},
		},
	},
	{
		Name: "first",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{
						fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)),
					}
				}
				if args[0].Type() != ARRAY_OBJ {
					return &Error{
						fmt.Sprintf("argument to `first` must be ARRAY, got %s", args[0].Type()),
					}
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}

				return nil
			},
		},
	},
	{
		Name: "last",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{
						fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)),
					}
				}
				if args[0].Type() != ARRAY_OBJ {
					return &Error{
						fmt.Sprintf("argument to `last` must be ARRAY, got %s", args[0].Type()),
					}
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)
				if length > 0 {
					return arr.Elements[length-1]
				}

				return nil
			},
		},
	},
	{
		Name: "rest",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{
						fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)),
					}
				}
				if args[0].Type() != ARRAY_OBJ {
					return &Error{
						fmt.Sprintf("argument to `rest` must be ARRAY, got %s", args[0].Type()),
					}
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)
				if length > 0 {
					newElements := make([]Object, length-1)
					copy(newElements, arr.Elements[1:length])
					return &Array{Elements: newElements}
				}

				return nil
			},
		},
	},
	{
		Name: "push",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return &Error{
						fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args)),
					}
				}
				if args[0].Type() != ARRAY_OBJ {
					return &Error{
						fmt.Sprintf("argument to `push` must be ARRAY, got %s", args[0].Type()),
					}
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)

				newElements := make([]Object, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]

				return &Array{Elements: newElements}
			},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, definition := range Builtins {
		if definition.Name == name {
			return definition.Builtin
		}
	}

	return nil
}
