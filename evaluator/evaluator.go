package evaluator

import (
	"MyInterpreter/ast"
	"MyInterpreter/object"
	"MyInterpreter/packages/mymath"
	"fmt"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	VOID  = &object.Void{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatements(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.CompoundAssignment:
		value := Eval(node.Value, env).(*object.Integer)
		if isError(value) {
			return value
		}

		val, ok := env.Get(node.Variable.String())
		if !ok {
			fmt.Println("Error")
			return value
		}

		switch node.Operator {
		case "+=":
			env.Set(node.Variable.String(), &object.Integer{Value: val.(*object.Integer).Value + value.Value})
		case "-=":
			env.Set(node.Variable.String(), &object.Integer{Value: val.(*object.Integer).Value - value.Value})
		case "*=":
			env.Set(node.Variable.String(), &object.Integer{Value: val.(*object.Integer).Value * value.Value})
		case "/=":
			env.Set(node.Variable.String(), &object.Integer{Value: val.(*object.Integer).Value / value.Value})
		}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.WhileLoop:
		return evalWhileLoop(node, env)
	}

	return nil

}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func nativeBoolToBooleanObject(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE

	//Interesting small Optimization
	//Instead of creating an object to each boolean, it's easier to implement
	// a solution in a singleton style, where everytime we get a Boolean Expression
	// We return our already existing Boolean object, just a small detail
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case right.Type() == object.STRING_OBJ && left.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "*": //This does not handle int multiplication, this is responsible for multiplication between strings and integers
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "**":
		return &object.Integer{Value: int64(mymath.Exponentiate(leftVal, rightVal))}
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch operator {
	case "+":
		return &object.String{Value: left.Inspect() + right.Inspect()}
	case "*":
		if left.Type() == right.Type() {
			return newError("unknown operator: %s %s %s", left.Inspect(), operator, right.Inspect())
		}
		if left.Type() == object.INTEGER_OBJ && right.Type() == object.STRING_OBJ {
			var string string
			nloop := left.(*object.Integer).Value
			for range nloop {
				string += right.Inspect()
			}
			return &object.String{Value: string}
		} else if right.Type() == object.INTEGER_OBJ && left.Type() == object.STRING_OBJ {
			var string string
			nloop := right.(*object.Integer).Value
			for range nloop {
				string += left.Inspect()
			}
			return &object.String{Value: string}
		} else {
			return newError(" %s operator not supported between %s and %s", operator, left.Type(), right.Type())
		}

	default:
		return nil
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}

}

func isTruthy(condition object.Object) bool {
	switch condition {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalBlockStatements(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}

	}
	return result
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {

	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	//just like evalBlock but for func parameters

	for _, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {

	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args) // add parameters as local scope vars
		evaluated := Eval(fn.Body, extendedEnv)    //evaluate BlockStatement
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function %s", fn.Type())
	}

}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.ScopedEnv(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx]) //add parameters as local variables
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported for %s", left.Type())
	}
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashkey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashkey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func evalArrayIndexExpression(left, index object.Object) object.Object {
	array, ok := left.(*object.Array)
	if !ok {
		return newError("index operator not supported for %s", left.Type())
	}

	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("%s can't be used as index", index.Type())
	}

	max_index := int64(len(array.Elements) - 1)

	if idx.Value < 0 {
		idx.Value = max_index + idx.Value + 1
	}

	if idx.Value > max_index {
		return NULL
	}

	return array.Elements[idx.Value]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject, ok := hash.(*object.Hash)
	if !ok {
		return newError("unusable as hash index: %s", hash.Type())
	}

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalWhileLoop(node *ast.WhileLoop, env *object.Environment) object.Object {
	var WhileTrue object.Object
	var evaluated object.Object

	WhileTrue = Eval(node.Condition, env)

	if isError(WhileTrue) {
		return WhileTrue
	}

	for isTruthy(Eval(node.Condition, env)) {
		evaluated = Eval(node.Consequence, env)
	}

	return evaluated
}
