package evaluator

import (
	"MyInterpreter/object"
	"fmt"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {

			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to 'push' must be ARRAY got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			//everytime we push to an array, we use heap memory to create a copy of the original array + pushed element
			//I'm not entirely sure if this is the most memory efficient option, but it is the solution we have for now
			// Besides, the way the author implemented Push, we gotta create another var to store the newvalue
			// as the original array is not modifier, instead, a copy is created and returned in its place :/

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}

		},
	},

	"remove": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. remove accepts 1 argument, got=%d instead", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("remove built-in is only supported for arrays, got=%s", args[0].Type())
			}
			Array := args[0].(*object.Array)

			for idx, elem := range Array.Elements {
				if elem.Inspect() == args[1].Inspect() {
					Array.Elements[idx] = Array.Elements[len(Array.Elements)-1]
					Array.Elements = Array.Elements[:len(Array.Elements)-1]
					return Array
				}
			}
			return newError("argument to be removed must be in the Array")

		},
	},
	"print": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return VOID
		},
	},
}
